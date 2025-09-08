package jit

import (
	"monkey/object"
	"sync"
	"unsafe"
)

// JITExtension はCompiledFunctionのJIT拡張情報
type JITExtension struct {
	// JIT実行情報
	CallCount   int64          // 呼び出し回数
	JITCompiled bool           // JITコンパイル済みフラグ
	NativeCode  *NativeCode    // ネイティブコード
	EntryPoint  unsafe.Pointer // エントリーポイント

	// プロファイリング情報
	LastCallTime int64 // 最後の呼び出し時刻（unix nano）
	TotalCalls   int64 // 総呼び出し回数

	// 最適化情報
	IsHotFunction     bool // ホット関数フラグ
	OptimizationLevel int  // 最適化レベル

	// 同期
	mutex sync.RWMutex
}

// JITRegistry はCompiledFunctionとJIT拡張の関連付けを管理
type JITRegistry struct {
	extensions map[*object.CompiledFunction]*JITExtension
	mutex      sync.RWMutex
}

// globalRegistry はグローバルなJIT拡張レジストリ
var globalRegistry = &JITRegistry{
	extensions: make(map[*object.CompiledFunction]*JITExtension),
}

// GetJITExtension は関数のJIT拡張を取得（なければ作成）
func GetJITExtension(fn *object.CompiledFunction) *JITExtension {
	return globalRegistry.GetExtension(fn)
}

// GetExtension は関数のJIT拡張を取得（なければ作成）
func (r *JITRegistry) GetExtension(fn *object.CompiledFunction) *JITExtension {
	r.mutex.RLock()
	ext, exists := r.extensions[fn]
	r.mutex.RUnlock()

	if exists {
		return ext
	}

	// 新しい拡張を作成
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// ダブルチェック（他のゴルーチンが作成している可能性）
	if ext, exists := r.extensions[fn]; exists {
		return ext
	}

	ext = &JITExtension{
		CallCount:         0,
		JITCompiled:       false,
		NativeCode:        nil,
		EntryPoint:        nil,
		LastCallTime:      0,
		TotalCalls:        0,
		IsHotFunction:     false,
		OptimizationLevel: 0,
	}

	r.extensions[fn] = ext
	return ext
}

// SetNativeCode はネイティブコードを設定
func (ext *JITExtension) SetNativeCode(code *NativeCode) {
	ext.mutex.Lock()
	defer ext.mutex.Unlock()

	ext.NativeCode = code
	ext.EntryPoint = code.EntryPoint
	ext.JITCompiled = true
}

// IncrementCallCount は呼び出し回数をインクリメント
func (ext *JITExtension) IncrementCallCount() int64 {
	ext.mutex.Lock()
	defer ext.mutex.Unlock()

	ext.CallCount++
	ext.TotalCalls++
	return ext.CallCount
}

// GetCallCount は呼び出し回数を取得
func (ext *JITExtension) GetCallCount() int64 {
	ext.mutex.RLock()
	defer ext.mutex.RUnlock()
	return ext.CallCount
}

// IsJITCompiled はJITコンパイル済みかを確認
func (ext *JITExtension) IsJITCompiled() bool {
	ext.mutex.RLock()
	defer ext.mutex.RUnlock()
	return ext.JITCompiled
}

// GetEntryPoint はエントリーポイントを取得
func (ext *JITExtension) GetEntryPoint() unsafe.Pointer {
	ext.mutex.RLock()
	defer ext.mutex.RUnlock()
	return ext.EntryPoint
}

// SetHotFunction はホット関数フラグを設定
func (ext *JITExtension) SetHotFunction(isHot bool) {
	ext.mutex.Lock()
	defer ext.mutex.Unlock()
	ext.IsHotFunction = isHot
}

// IsHot はホット関数かを確認
func (ext *JITExtension) IsHot() bool {
	ext.mutex.RLock()
	defer ext.mutex.RUnlock()
	return ext.IsHotFunction
}

// GetStats は統計情報を取得
func (ext *JITExtension) GetStats() map[string]interface{} {
	ext.mutex.RLock()
	defer ext.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["call_count"] = ext.CallCount
	stats["total_calls"] = ext.TotalCalls
	stats["jit_compiled"] = ext.JITCompiled
	stats["is_hot"] = ext.IsHotFunction
	stats["optimization_level"] = ext.OptimizationLevel
	stats["last_call_time"] = ext.LastCallTime

	if ext.NativeCode != nil {
		stats["native_code_size"] = ext.NativeCode.Size
		stats["compile_time_ns"] = ext.NativeCode.CompileTime
	}

	return stats
}

// GetAllExtensions は全てのJIT拡張を取得
func (r *JITRegistry) GetAllExtensions() map[*object.CompiledFunction]*JITExtension {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// コピーを作成
	result := make(map[*object.CompiledFunction]*JITExtension)
	for fn, ext := range r.extensions {
		result[fn] = ext
	}

	return result
}

// GetHotFunctions はホット関数のリストを取得
func (r *JITRegistry) GetHotFunctions() []*object.CompiledFunction {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var hotFunctions []*object.CompiledFunction
	for fn, ext := range r.extensions {
		if ext.IsHot() {
			hotFunctions = append(hotFunctions, fn)
		}
	}

	return hotFunctions
}

// GetCompiledFunctions はJITコンパイル済み関数のリストを取得
func (r *JITRegistry) GetCompiledFunctions() []*object.CompiledFunction {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var compiledFunctions []*object.CompiledFunction
	for fn, ext := range r.extensions {
		if ext.IsJITCompiled() {
			compiledFunctions = append(compiledFunctions, fn)
		}
	}

	return compiledFunctions
}

// Reset はレジストリをリセット
func (r *JITRegistry) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.extensions = make(map[*object.CompiledFunction]*JITExtension)
}

// GetGlobalRegistry はグローバルレジストリを取得
func GetGlobalRegistry() *JITRegistry {
	return globalRegistry
}
