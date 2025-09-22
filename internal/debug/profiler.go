package debug

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// Profiler 性能分析器
type Profiler struct {
	server *http.Server
	port   string
}

// NewProfiler 创建新的性能分析器
func NewProfiler(port string) *Profiler {
	if port == "" {
		port = "6060"
	}

	return &Profiler{
		port: port,
	}
}

// Start 启动性能分析服务器
func (p *Profiler) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Java Analyzer Profiler</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .link { display: block; margin: 10px 0; padding: 10px; background: #f0f0f0; text-decoration: none; color: #333; border-radius: 4px; }
        .link:hover { background: #e0e0e0; }
    </style>
</head>
<body>
    <h1>Java Analyzer 性能分析</h1>
    <a href="/debug/pprof/" class="link">Go pprof 主页</a>
    <a href="/debug/pprof/profile" class="link">CPU Profile (30秒)</a>
    <a href="/debug/pprof/heap" class="link">内存堆分析</a>
    <a href="/debug/pprof/goroutine" class="link">Goroutine 分析</a>
    <a href="/debug/pprof/block" class="link">阻塞分析</a>
    <a href="/debug/pprof/mutex" class="link">互斥锁分析</a>
    <a href="/debug/pprof/trace" class="link">执行跟踪</a>
    <a href="/stats" class="link">运行时统计</a>
</body>
</html>
		`)
	})

	mux.HandleFunc("/stats", p.handleStats)

	p.server = &http.Server{
		Addr:    ":" + p.port,
		Handler: mux,
	}

	go func() {
		Info("启动性能分析服务器: http://localhost:%s", p.port)
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Error("性能分析服务器启动失败: %v", err)
		}
	}()

	return nil
}

// Stop 停止性能分析服务器
func (p *Profiler) Stop() error {
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return p.server.Shutdown(ctx)
	}
	return nil
}

// handleStats 处理运行时统计
func (p *Profiler) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Fprintf(w, "Java Analyzer 运行时统计\n")
	fmt.Fprintf(w, "========================\n\n")

	fmt.Fprintf(w, "内存使用:\n")
	fmt.Fprintf(w, "  分配的内存: %d KB\n", m.Alloc/1024)
	fmt.Fprintf(w, "  总分配的内存: %d KB\n", m.TotalAlloc/1024)
	fmt.Fprintf(w, "  系统内存: %d KB\n", m.Sys/1024)
	fmt.Fprintf(w, "  堆内存: %d KB\n", m.HeapAlloc/1024)
	fmt.Fprintf(w, "  堆系统内存: %d KB\n", m.HeapSys/1024)
	fmt.Fprintf(w, "  堆空闲内存: %d KB\n", m.HeapIdle/1024)
	fmt.Fprintf(w, "  堆使用内存: %d KB\n", m.HeapInuse/1024)
	fmt.Fprintf(w, "  堆释放内存: %d KB\n", m.HeapReleased/1024)
	fmt.Fprintf(w, "  堆对象数: %d\n", m.HeapObjects)

	fmt.Fprintf(w, "\n垃圾回收:\n")
	fmt.Fprintf(w, "  GC次数: %d\n", m.NumGC)
	fmt.Fprintf(w, "  上次GC时间: %s\n", time.Unix(0, int64(m.LastGC)).Format(time.RFC3339))
	fmt.Fprintf(w, "  GC暂停时间: %d ns\n", m.PauseTotalNs)

	fmt.Fprintf(w, "\nGoroutine:\n")
	fmt.Fprintf(w, "  当前Goroutine数: %d\n", runtime.NumGoroutine())

	fmt.Fprintf(w, "\nCPU:\n")
	fmt.Fprintf(w, "  CPU核心数: %d\n", runtime.NumCPU())
	fmt.Fprintf(w, "  GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
}

// StartWithSignal 启动性能分析服务器并监听信号
func (p *Profiler) StartWithSignal() error {
	if err := p.Start(); err != nil {
		return err
	}

	// 监听信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	Info("收到停止信号，关闭性能分析服务器...")

	return p.Stop()
}

// 全局性能分析器
var globalProfiler *Profiler

// StartProfiler 启动全局性能分析器
func StartProfiler(port string) error {
	globalProfiler = NewProfiler(port)
	return globalProfiler.Start()
}

// StopProfiler 停止全局性能分析器
func StopProfiler() error {
	if globalProfiler != nil {
		return globalProfiler.Stop()
	}
	return nil
}

// StartProfilerWithSignal 启动全局性能分析器并监听信号
func StartProfilerWithSignal(port string) error {
	globalProfiler = NewProfiler(port)
	return globalProfiler.StartWithSignal()
}
