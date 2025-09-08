package jit

import (
	"fmt"
	"monkey/object"
	"sync"
	"unsafe"
)

// JITThreshold は関数をJITコンパイルする閾値
const JITThreshold = 100

// JIT は Just-In-Time コンパイラのメイン構造体
type JIT struct {
	// プロファイラー
	profiler *Profiler

	// コンパイラー
	compiler *Compiler

	// ネイティブコードキャッシュ
	codeCache  map[*object.CompiledFunction]*NativeCode
	cacheMutex sync.RWMutex

	// メモリ管理
	memoryManager *MemoryManager

	// 統計情報
	stats *Statistics

	// 設定
	config *Config
}

// Config はJITコンパイラーの設定
type Config struct {
	Threshold         int  // JITコンパイル閾値
	OptimizationLevel int  // 最適化レベル (0-3)
	EnableProfiling   bool // プロファイリング有効化
	MaxCodeCacheSize  int  // コードキャッシュ最大サイズ (bytes)
}

// Statistics はJIT統計情報
type Statistics struct {
	FunctionsCompiled int64 // コンパイルされた関数数
	TotalCompileTime  int64 // 総コンパイル時間 (ns)
	CacheHits         int64 // キャッシュヒット数
	CacheMisses       int64 // キャッシュミス数
	CodeCacheSize     int64 // コードキャッシュサイズ
	MemoryUsage       int64 // メモリ使用量
}

// NativeCode はコンパイルされたネイティブコードを表す
type NativeCode struct {
	Function    *object.CompiledFunction // 元の関数
	Code        []byte                   // ネイティブコード
	Size        int                      // コードサイズ
	EntryPoint  unsafe.Pointer           // エントリーポイント
	Metadata    *Metadata                // メタデータ
	CompileTime int64                    // コンパイル時間 (ns)
}

// Metadata はコンパイルされた関数のメタデータ
type Metadata struct {
	OriginalSize     int            // 元のバイトコードサイズ
	OptimizationInfo string         // 最適化情報
	TypeProfile      map[int]string // 型プロファイル
}

// NewJIT は新しいJITインスタンスを作成
func NewJIT(config *Config) *JIT {
	if config == nil {
		config = &Config{
			Threshold:         JITThreshold,
			OptimizationLevel: 1,
			EnableProfiling:   true,
			MaxCodeCacheSize:  10 * 1024 * 1024, // 10MB
		}
	}

	memManager := NewMemoryManager(config.MaxCodeCacheSize)
	compiler := NewCompiler()
	compiler.SetMemoryManager(memManager)

	jit := &JIT{
		profiler:      NewProfiler(),
		compiler:      compiler,
		codeCache:     make(map[*object.CompiledFunction]*NativeCode),
		memoryManager: memManager,
		stats:         &Statistics{},
		config:        config,
	}

	return jit
}

// ShouldCompile は関数をJITコンパイルすべきかを判定
func (j *JIT) ShouldCompile(fn *object.CompiledFunction) bool {
	if j.config == nil || !j.config.EnableProfiling {
		return false
	}

	// すでにコンパイル済みか確認
	j.cacheMutex.RLock()
	_, exists := j.codeCache[fn]
	j.cacheMutex.RUnlock()

	if exists {
		return false
	}

	// プロファイラーで判定
	return j.profiler.ShouldCompile(fn, j.config.Threshold)
}

// GetNativeCode はコンパイル済みネイティブコードを取得
func (j *JIT) GetNativeCode(fn *object.CompiledFunction) (*NativeCode, bool) {
	j.cacheMutex.RLock()
	defer j.cacheMutex.RUnlock()

	code, exists := j.codeCache[fn]
	if exists {
		j.stats.CacheHits++
	} else {
		j.stats.CacheMisses++
	}

	return code, exists
}

// CompileFunction は関数をJITコンパイル
func (j *JIT) CompileFunction(fn *object.CompiledFunction) (*NativeCode, error) {
	// すでにコンパイル済みか確認
	if code, exists := j.GetNativeCode(fn); exists {
		return code, nil
	}

	// コンパイル実行
	nativeCode, err := j.compiler.Compile(fn, j.config.OptimizationLevel)
	if err != nil {
		return nil, fmt.Errorf("JIT compilation failed: %w", err)
	}

	// キャッシュに保存
	j.cacheMutex.Lock()
	j.codeCache[fn] = nativeCode
	j.stats.FunctionsCompiled++
	j.stats.TotalCompileTime += nativeCode.CompileTime
	j.stats.CodeCacheSize += int64(nativeCode.Size)
	j.cacheMutex.Unlock()

	return nativeCode, nil
}

// RecordFunctionCall は関数呼び出しを記録
func (j *JIT) RecordFunctionCall(fn *object.CompiledFunction, args []object.Object) {
	if j.config.EnableProfiling {
		j.profiler.RecordCall(fn, args)
	}
}

// GetStatistics は統計情報を取得
func (j *JIT) GetStatistics() Statistics {
	j.cacheMutex.RLock()
	defer j.cacheMutex.RUnlock()

	stats := *j.stats
	stats.MemoryUsage = j.memoryManager.GetUsage()
	return stats
}

// PrintStatistics は統計情報を出力
func (j *JIT) PrintStatistics() {
	stats := j.GetStatistics()
	fmt.Printf("JIT Statistics:\n")
	fmt.Printf("  Functions compiled: %d\n", stats.FunctionsCompiled)
	fmt.Printf("  Total compile time: %.2f ms\n", float64(stats.TotalCompileTime)/1e6)
	fmt.Printf("  Cache hits: %d\n", stats.CacheHits)
	fmt.Printf("  Cache misses: %d\n", stats.CacheMisses)
	fmt.Printf("  Code cache size: %.2f KB\n", float64(stats.CodeCacheSize)/1024)
	fmt.Printf("  Memory usage: %.2f KB\n", float64(stats.MemoryUsage)/1024)

	if stats.CacheHits+stats.CacheMisses > 0 {
		hitRate := float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses) * 100
		fmt.Printf("  Cache hit rate: %.1f%%\n", hitRate)
	}
}

// Cleanup はJITリソースをクリーンアップ
func (j *JIT) Cleanup() error {
	// メモリ管理のクリーンアップ
	return j.memoryManager.Cleanup()
}
