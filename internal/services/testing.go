package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	models "gibraltar/internal/models"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

type URLTestService struct {
	URL string
}

func NewVlessTestService(url string) *URLTestService {
	return &URLTestService{
		URL: url,
	}
}

func (t *URLTestService) Test(config *models.VlessConfig, localPort int) (time.Duration, error) {
	outbound := buildVlessOutbound(config)
	singBoxConfig := buildSingBoxConfig(outbound, localPort)
	pattern := fmt.Sprintf("singbox-config-*-%s.json", config.Server)
	tmp, err := os.CreateTemp("", pattern)
	if err != nil {
		return 0 * time.Millisecond, fmt.Errorf("create temp config: %w", err)
	}
	defer os.Remove(tmp.Name())
	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(singBoxConfig); err != nil {
		tmp.Close()
		return 0, fmt.Errorf("write config: %w", err)
	}
	tmp.Close()
	cmd := exec.Command("sing-box", "run", "-c", tmp.Name())

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start sing-box: %w", err)
	}
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	started := true
	defer func() {
		if started && cmd.Process != nil {
			_ = cmd.Process.Kill()
			<-done
		}
	}()

	stringPort := strconv.Itoa(localPort)

	if err := waitPort(net.JoinHostPort("127.0.0.1", stringPort), 5*time.Second); err != nil {
		return 0, fmt.Errorf("sing-box inbound not ready: %w", err)
	}

	dialer, err := proxy.SOCKS5("tcp", net.JoinHostPort("127.0.0.1", stringPort), nil, proxy.Direct)
	if err != nil {
		return 0, fmt.Errorf("create socks5 dialer: %w", err)
	}
	var latency time.Duration

	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	switch config.Type {
	case "grpc":
		trace := &httptrace.ClientTrace{}
		req, err := http.NewRequest("GET", "https://www.cloudflare.com/cdn-cgi/trace", nil)
		if err != nil {
			return 0, err
		}
		req = req.WithContext(httptrace.WithClientTrace(context.TODO(), trace))
		for i := 0; i < 2; i++ {
			start := time.Now()
			resp, err := client.Do(req)
			latency = max(latency, time.Since(start))
			if err != nil {
				return 0, err
			}
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}

		}
	default:
		start := time.Now()
		resp, err := client.Get(t.URL)
		latency = time.Since(start)
		if err != nil {
			return 0, fmt.Errorf("probe failed: %w", err)
		}
		_ = resp.Body.Close()
	}

	started = false
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
		<-done
	}

	return latency, nil
}

func waitPort(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 400*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", addr)
}

func TLSTest(address, port, serverName string, timeout time.Duration) (time.Duration, error) {
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	start := time.Now()

	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(address, port), &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: true,
	})
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	return time.Since(start), nil
}

func (t *URLTestService) TestConfigs(configs []*models.VlessConfig, workersCount int) {
	log.Printf("%d configs will be tested\n", len(configs))
	ports := make([]int, 0, workersCount)
	for i := 0; i < workersCount; i++ {
		ports = append(ports, 2081+i)
	}
	jobs := make(chan int)

	var wg sync.WaitGroup
	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		port := ports[i]
		go func(p int) {
			defer wg.Done()
			for index := range jobs {
				t.setTestResultValue(configs[index], port)
			}
		}(port)
	}

	for i := range configs {
		jobs <- i
	}

	close(jobs)
	wg.Wait()
}

func (t *URLTestService) setTestResultValue(config *models.VlessConfig, localPort int) {
	time, err := TLSTest(config.Server, strconv.Itoa(config.Port), config.SNI, 2*time.Second)
	if time <= 0 || err != nil {
		config.TestResult = -1
		config.Stability = onFailure(config.Stability)
		return
	}
	lat, err := t.Test(config, localPort)
	if err != nil {
		config.TestResult = -1
		config.Stability = onFailure(config.Stability)
	} else {
		config.TestResult = int(lat.Milliseconds())
		config.Stability = onSuccess(config.Stability)
	}
}
