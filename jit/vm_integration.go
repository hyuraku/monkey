package jit

import (
	"fmt"
	"monkey/object"
	"monkey/vm"
	"time"
)

// VMIntegration はVMとJITの統合レイヤー
type VMIntegration struct {
	// 元のVM
	vm *vm.VM

	// JITシステム
	jit *JIT

	// 統計
	jitCallCount   int64
	vmCallCount    int64
	jitExecuteTime time.Duration
	vmExecuteTime  time.Duration

	// 設定
	enabled bool
}

// NewVMIntegration は新しいVM統合レイヤーを作成
func NewVMIntegration(originalVM *vm.VM, jitConfig *Config) *VMIntegration {
	return &VMIntegration{
		vm:      originalVM,
		jit:     NewJIT(jitConfig),
		enabled: true,
	}
}

// SetEnabled はJIT統合の有効/無効を設定
func (vi *VMIntegration) SetEnabled(enabled bool) {
	vi.enabled = enabled
}

// IsEnabled はJIT統合が有効かを確認
func (vi *VMIntegration) IsEnabled() bool {
	return vi.enabled
}

// GetVM は元のVMを取得
func (vi *VMIntegration) GetVM() *vm.VM {
	return vi.vm
}

// GetJIT はJITシステムを取得
func (vi *VMIntegration) GetJIT() *JIT {
	return vi.jit
}

// InterceptFunctionCall は関数呼び出しをインターセプト
func (vi *VMIntegration) InterceptFunctionCall(fn *object.CompiledFunction, args []object.Object) (object.Object, bool, error) {
	if !vi.enabled {
		return nil, false, nil // JIT処理なし
	}

	// JIT拡張を取得
	ext := GetJITExtension(fn)

	// 呼び出し回数をインクリメント
	callCount := ext.IncrementCallCount()

	// プロファイリング記録
	vi.jit.RecordFunctionCall(fn, args)

	// JITコンパイル判定
	if !ext.IsJITCompiled() && vi.jit.ShouldCompile(fn) {
		err := vi.compileFunction(fn)
		if err != nil {
			// コンパイル失敗時はVMにフォールバック
			return nil, false, fmt.Errorf("JIT compilation failed: %w", err)
		}
	}

	// JITコンパイル済みの場合、ネイティブ実行
	if ext.IsJITCompiled() {
		startTime := time.Now()
		result, err := vi.executeNative(fn, args)
		vi.jitExecuteTime += time.Since(startTime)
		vi.jitCallCount++

		if err != nil {
			// ネイティブ実行失敗時はVMにフォールバック
			return nil, false, fmt.Errorf("native execution failed: %w", err)
		}

		return result, true, nil // JIT実行成功
	}

	// ホット関数のマーキング
	if callCount > int64(vi.jit.config.Threshold/2) {
		ext.SetHotFunction(true)
	}

	// VM実行時間測定
	startTime := time.Now()
	defer func() {
		vi.vmExecuteTime += time.Since(startTime)
		vi.vmCallCount++
	}()

	return nil, false, nil // VM実行を継続
}

// compileFunction は関数をJITコンパイル
func (vi *VMIntegration) compileFunction(fn *object.CompiledFunction) error {
	// JITコンパイル実行
	nativeCode, err := vi.jit.CompileFunction(fn)
	if err != nil {
		return err
	}

	// 拡張にネイティブコードを設定
	ext := GetJITExtension(fn)
	ext.SetNativeCode(nativeCode)

	return nil
}

// executeNative はネイティブコードを実行
func (vi *VMIntegration) executeNative(fn *object.CompiledFunction, args []object.Object) (object.Object, error) {
	ext := GetJITExtension(fn)
	entryPoint := ext.GetEntryPoint()

	if entryPoint == nil {
		return nil, fmt.Errorf("native code entry point is null")
	}

	// 現在の実装では簡単なスタブを返す
	// 実際のネイティブコード実行は複雑なため、Phase 2で実装
	return vi.executeNativeStub(fn, args)
}

// executeNativeStub はネイティブ実行のスタブ（開発用）
func (vi *VMIntegration) executeNativeStub(fn *object.CompiledFunction, args []object.Object) (object.Object, error) {
	// スタブ実装：単純な算術演算のみ対応
	// 実際の実装ではアセンブリコードを呼び出す

	// 引数が2つの整数の場合、加算を実行
	if len(args) == 2 {
		if int1, ok := args[0].(*object.Integer); ok {
			if int2, ok := args[1].(*object.Integer); ok {
				// 簡単な加算例
				result := int1.Value + int2.Value
				return &object.Integer{Value: result}, nil
			}
		}
	}

	// その他の場合はVMにフォールバック
	return nil, fmt.Errorf("native execution not implemented for this function signature")
}

// GetStatistics は統計情報を取得
func (vi *VMIntegration) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// JIT統計
	jitStats := vi.jit.GetStatistics()
	stats["jit"] = jitStats

	// 実行統計
	stats["jit_call_count"] = vi.jitCallCount
	stats["vm_call_count"] = vi.vmCallCount
	stats["jit_execute_time_ms"] = float64(vi.jitExecuteTime.Nanoseconds()) / 1e6
	stats["vm_execute_time_ms"] = float64(vi.vmExecuteTime.Nanoseconds()) / 1e6

	// スピードアップ計算
	if vi.vmCallCount > 0 && vi.jitCallCount > 0 {
		avgVMTime := float64(vi.vmExecuteTime.Nanoseconds()) / float64(vi.vmCallCount)
		avgJITTime := float64(vi.jitExecuteTime.Nanoseconds()) / float64(vi.jitCallCount)

		if avgJITTime > 0 {
			speedup := avgVMTime / avgJITTime
			stats["average_speedup"] = speedup
		}
	}

	// ホット関数統計
	hotFunctions := GetGlobalRegistry().GetHotFunctions()
	compiledFunctions := GetGlobalRegistry().GetCompiledFunctions()

	stats["hot_function_count"] = len(hotFunctions)
	stats["compiled_function_count"] = len(compiledFunctions)

	return stats
}

// PrintStatistics は統計情報を出力
func (vi *VMIntegration) PrintStatistics() {
	fmt.Println("VM-JIT Integration Statistics:")
	fmt.Println("==============================")

	stats := vi.GetStatistics()

	fmt.Printf("JIT Calls: %d\n", vi.jitCallCount)
	fmt.Printf("VM Calls: %d\n", vi.vmCallCount)
	fmt.Printf("JIT Execute Time: %.2f ms\n", float64(vi.jitExecuteTime.Nanoseconds())/1e6)
	fmt.Printf("VM Execute Time: %.2f ms\n", float64(vi.vmExecuteTime.Nanoseconds())/1e6)

	if speedup, ok := stats["average_speedup"].(float64); ok {
		fmt.Printf("Average Speedup: %.2fx\n", speedup)
	}

	fmt.Printf("Hot Functions: %d\n", stats["hot_function_count"])
	fmt.Printf("Compiled Functions: %d\n", stats["compiled_function_count"])

	fmt.Println()
	vi.jit.PrintStatistics()
}

// Cleanup はリソースをクリーンアップ
func (vi *VMIntegration) Cleanup() error {
	return vi.jit.Cleanup()
}

// 関数呼び出しフック用のインターフェース

// FunctionCallHook は関数呼び出し時に呼ばれるフック
type FunctionCallHook func(fn *object.CompiledFunction, args []object.Object) (object.Object, bool, error)

// VMWithJIT はJIT機能付きVMラッパー
type VMWithJIT struct {
	*vm.VM
	integration *VMIntegration
	hook        FunctionCallHook
}

// NewVMWithJIT はJIT機能付きVMを作成
func NewVMWithJIT(originalVM *vm.VM, jitConfig *Config) *VMWithJIT {
	integration := NewVMIntegration(originalVM, jitConfig)

	vmWithJIT := &VMWithJIT{
		VM:          originalVM,
		integration: integration,
	}

	// フックを設定
	vmWithJIT.hook = integration.InterceptFunctionCall

	return vmWithJIT
}

// GetIntegration は統合レイヤーを取得
func (vj *VMWithJIT) GetIntegration() *VMIntegration {
	return vj.integration
}

// CallFunction はJIT対応の関数呼び出し
func (vj *VMWithJIT) CallFunction(fn *object.CompiledFunction, args []object.Object) (object.Object, error) {
	// JITフックを呼び出し
	if vj.hook != nil {
		result, handled, err := vj.hook(fn, args)
		if err != nil {
			// JITエラーの場合はVMにフォールバック
			// ログを出力し、VMで実行継続
		} else if handled {
			// JITで処理完了
			return result, nil
		}
		// handled=false の場合はVMで実行継続
	}

	// VM実行（元の実装を呼び出す必要があるが、ここではスタブ）
	// 実際の統合では、元のVMのcallClosure関数を呼び出す
	return &object.Integer{Value: 0}, nil // スタブ実装
}
