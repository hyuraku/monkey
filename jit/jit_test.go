package jit

import (
	"monkey/code"
	"monkey/object"
	"testing"
)

func TestJITBasicFlow(t *testing.T) {
	// JIT設定
	config := &Config{
		Threshold:         5, // 低い閾値でテスト
		OptimizationLevel: 1,
		EnableProfiling:   true,
		MaxCodeCacheSize:  1024 * 1024, // 1MB
	}

	// JITインスタンス作成
	jit := NewJIT(config)
	defer jit.Cleanup()

	// テスト用のCompiledFunction作成（簡単な加算）
	instructions := []byte{
		byte(code.OpConstant), 0, 0, // load constant 0
		byte(code.OpConstant), 0, 1, // load constant 1
		byte(code.OpAdd),         // add
		byte(code.OpReturnValue), // return
	}

	fn := &object.CompiledFunction{
		Instructions:  instructions,
		NumLocals:     0,
		NumParameters: 0,
	}

	// 初期状態：JITコンパイルされていない
	if jit.ShouldCompile(fn) {
		t.Error("Function should not be compiled initially")
	}

	// 閾値まで呼び出しを記録
	args := []object.Object{
		&object.Integer{Value: 1},
		&object.Integer{Value: 2},
	}

	for i := 0; i < 5; i++ {
		jit.RecordFunctionCall(fn, args)
	}

	// 閾値到達：JITコンパイルが必要
	if !jit.ShouldCompile(fn) {
		t.Error("Function should be compiled after threshold")
	}

	// JITコンパイル実行
	nativeCode, err := jit.CompileFunction(fn)
	if err != nil {
		t.Errorf("JIT compilation failed: %v", err)
	}

	if nativeCode == nil {
		t.Error("Native code should not be nil")
	}

	// キャッシュから取得
	cachedCode, exists := jit.GetNativeCode(fn)
	if !exists {
		t.Error("Compiled function should be cached")
	}

	if cachedCode != nativeCode {
		t.Error("Cached code should match compiled code")
	}

	// 統計確認
	stats := jit.GetStatistics()
	if stats.FunctionsCompiled != 1 {
		t.Errorf("Expected 1 compiled function, got %d", stats.FunctionsCompiled)
	}
}

func TestProfiler(t *testing.T) {
	profiler := NewProfiler()

	// テスト関数
	fn := &object.CompiledFunction{
		Instructions:  []byte{byte(code.OpReturn)},
		NumLocals:     0,
		NumParameters: 2,
	}

	// 呼び出し記録
	args := []object.Object{
		&object.Integer{Value: 10},
		&object.String{Value: "test"},
	}

	for i := 0; i < 10; i++ {
		profiler.RecordCall(fn, args)
	}

	// プロファイル確認
	profile, exists := profiler.GetProfile(fn)
	if !exists {
		t.Error("Profile should exist")
	}

	if profile.CallCount != 10 {
		t.Errorf("Expected 10 calls, got %d", profile.CallCount)
	}

	// 引数型の確認
	argType, count := profiler.GetDominantArgumentType(fn, 0)
	if argType != "INTEGER" {
		t.Errorf("Expected INTEGER type, got %s", argType)
	}
	if count != 10 {
		t.Errorf("Expected 10 INTEGER arguments, got %d", count)
	}

	// 戻り値記録
	result := &object.Integer{Value: 42}
	for i := 0; i < 5; i++ {
		profiler.RecordReturn(fn, result)
	}

	returnType, returnCount := profiler.GetDominantReturnType(fn)
	if returnType != "INTEGER" {
		t.Errorf("Expected INTEGER return type, got %s", returnType)
	}
	if returnCount != 5 {
		t.Errorf("Expected 5 INTEGER returns, got %d", returnCount)
	}

	// ホット関数リスト
	hotFunctions := profiler.GetHotFunctions(5)
	if len(hotFunctions) != 1 {
		t.Errorf("Expected 1 hot function, got %d", len(hotFunctions))
	}
}

func TestMemoryManager(t *testing.T) {
	mm := NewMemoryManager(1024 * 1024) // 1MB
	defer mm.Cleanup()

	// メモリ確保
	ptr, err := mm.AllocateExecutable(1024)
	if err != nil {
		t.Errorf("Memory allocation failed: %v", err)
	}

	if ptr == nil {
		t.Error("Allocated pointer should not be nil")
	}

	// コード書き込み
	testCode := []byte{0x90, 0x90, 0x90} // NOP instructions
	err = mm.WriteCode(ptr, testCode)
	if err != nil {
		t.Errorf("Code writing failed: %v", err)
	}

	// 実行可能設定
	err = mm.MakeExecutable(ptr, len(testCode))
	if err != nil {
		t.Errorf("Making executable failed: %v", err)
	}

	// 使用量確認
	usage := mm.GetUsage()
	if usage < int64(len(testCode)) {
		t.Errorf("Usage should be at least %d, got %d", len(testCode), usage)
	}

	// メモリ解放
	err = mm.Free(ptr)
	if err != nil {
		t.Errorf("Memory free failed: %v", err)
	}
}

func TestJITExtensions(t *testing.T) {
	// テスト関数
	fn := &object.CompiledFunction{
		Instructions:  []byte{byte(code.OpReturn)},
		NumLocals:     0,
		NumParameters: 0,
	}

	// JIT拡張取得
	ext := GetJITExtension(fn)
	if ext == nil {
		t.Error("JIT extension should not be nil")
	}

	// 初期状態確認
	if ext.IsJITCompiled() {
		t.Error("Function should not be compiled initially")
	}

	if ext.GetCallCount() != 0 {
		t.Error("Initial call count should be 0")
	}

	// 呼び出し回数インクリメント
	count := ext.IncrementCallCount()
	if count != 1 {
		t.Errorf("Expected call count 1, got %d", count)
	}

	// ホット関数設定
	ext.SetHotFunction(true)
	if !ext.IsHot() {
		t.Error("Function should be marked as hot")
	}

	// 統計取得
	stats := ext.GetStats()
	if stats["call_count"] != int64(1) {
		t.Errorf("Expected call count 1 in stats, got %v", stats["call_count"])
	}

	if stats["is_hot"] != true {
		t.Error("Function should be marked as hot in stats")
	}
}

func TestCompiler(t *testing.T) {
	mm := NewMemoryManager(1024 * 1024)
	defer mm.Cleanup()

	compiler := NewCompiler()
	compiler.SetMemoryManager(mm)

	// テスト関数
	instructions := []byte{
		byte(code.OpConstant), 0, 0,
		byte(code.OpConstant), 0, 1,
		byte(code.OpAdd),
		byte(code.OpReturnValue),
	}

	fn := &object.CompiledFunction{
		Instructions:  instructions,
		NumLocals:     0,
		NumParameters: 0,
	}

	// コンパイル実行
	nativeCode, err := compiler.Compile(fn, 1)
	if err != nil {
		t.Errorf("Compilation failed: %v", err)
	}

	if nativeCode == nil {
		t.Error("Native code should not be nil")
	}

	if nativeCode.Function != fn {
		t.Error("Native code should reference original function")
	}

	if len(nativeCode.Code) == 0 {
		t.Error("Native code should not be empty")
	}

	// メタデータ確認
	if nativeCode.Metadata == nil {
		t.Error("Metadata should not be nil")
	}

	if nativeCode.Metadata.OriginalSize != len(instructions) {
		t.Errorf("Expected original size %d, got %d", len(instructions), nativeCode.Metadata.OriginalSize)
	}

	// 統計確認
	stats := compiler.GetStatistics()
	if stats["compiled_functions"] != int64(1) {
		t.Errorf("Expected 1 compiled function, got %v", stats["compiled_functions"])
	}
}

func BenchmarkJITCompilation(b *testing.B) {
	config := &Config{
		Threshold:         1,
		OptimizationLevel: 1,
		EnableProfiling:   true,
		MaxCodeCacheSize:  10 * 1024 * 1024,
	}

	jit := NewJIT(config)
	defer jit.Cleanup()

	// テスト関数
	instructions := []byte{
		byte(code.OpConstant), 0, 0,
		byte(code.OpConstant), 0, 1,
		byte(code.OpAdd),
		byte(code.OpReturnValue),
	}

	args := []object.Object{
		&object.Integer{Value: 1},
		&object.Integer{Value: 2},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 新しい関数を作成（キャッシュ効果を避けるため）
		testFn := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     0,
			NumParameters: 0,
		}

		jit.RecordFunctionCall(testFn, args)
		if jit.ShouldCompile(testFn) {
			_, err := jit.CompileFunction(testFn)
			if err != nil {
				b.Errorf("Compilation failed: %v", err)
			}
		}
	}
}

func BenchmarkProfilerRecording(b *testing.B) {
	profiler := NewProfiler()

	fn := &object.CompiledFunction{
		Instructions:  []byte{byte(code.OpReturn)},
		NumLocals:     0,
		NumParameters: 2,
	}

	args := []object.Object{
		&object.Integer{Value: 42},
		&object.String{Value: "test"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		profiler.RecordCall(fn, args)
	}
}

// TestIntegrationFlow はJITシステム全体の統合テスト
func TestIntegrationFlow(t *testing.T) {
	config := &Config{
		Threshold:         3,
		OptimizationLevel: 1,
		EnableProfiling:   true,
		MaxCodeCacheSize:  1024 * 1024,
	}

	jit := NewJIT(config)
	defer jit.Cleanup()

	// 複数の関数でテスト
	functions := make([]*object.CompiledFunction, 3)
	for i := 0; i < 3; i++ {
		functions[i] = &object.CompiledFunction{
			Instructions:  []byte{byte(code.OpConstant), 0, byte(i), byte(code.OpReturnValue)},
			NumLocals:     0,
			NumParameters: 0,
		}
	}

	args := []object.Object{&object.Integer{Value: 1}}

	// すべての関数を閾値まで呼び出し
	for _, fn := range functions {
		for i := 0; i < 4; i++ { // 閾値(3)を超える
			jit.RecordFunctionCall(fn, args)
		}
	}

	// すべての関数がコンパイル対象になっているかチェック
	compiledCount := 0
	for _, fn := range functions {
		if jit.ShouldCompile(fn) {
			_, err := jit.CompileFunction(fn)
			if err != nil {
				t.Errorf("Failed to compile function: %v", err)
			}
			compiledCount++
		}
	}

	if compiledCount != 3 {
		t.Errorf("Expected 3 compiled functions, got %d", compiledCount)
	}

	// 統計確認
	stats := jit.GetStatistics()
	if stats.FunctionsCompiled != 3 {
		t.Errorf("Expected 3 compiled functions in stats, got %d", stats.FunctionsCompiled)
	}

	// 全体統計出力（デバッグ用）
	t.Logf("JIT Statistics: %+v", stats)
}
