package jit

import (
	"fmt"
	"monkey/code"
	"monkey/object"
	"time"
	"unsafe"
)

// Compiler はバイトコードをネイティブコードにコンパイルする
type Compiler struct {
	// メモリ管理への参照（外部から設定）
	memoryManager *MemoryManager

	// コンパイル統計
	compiledFunctions int64
	totalCompileTime  time.Duration
}

// CompilationContext はコンパイル時のコンテキスト
type CompilationContext struct {
	Function          *object.CompiledFunction
	OptimizationLevel int
	TypeHints         map[int]string // 引数のインデックス -> 推定型
}

// NewCompiler は新しいコンパイラーを作成
func NewCompiler() *Compiler {
	return &Compiler{
		compiledFunctions: 0,
	}
}

// SetMemoryManager はメモリマネージャーを設定
func (c *Compiler) SetMemoryManager(mm *MemoryManager) {
	c.memoryManager = mm
}

// Compile は関数をネイティブコードにコンパイル
func (c *Compiler) Compile(fn *object.CompiledFunction, optimizationLevel int) (*NativeCode, error) {
	startTime := time.Now()

	if c.memoryManager == nil {
		return nil, fmt.Errorf("memory manager not set")
	}

	// コンパイルコンテキストを作成
	ctx := &CompilationContext{
		Function:          fn,
		OptimizationLevel: optimizationLevel,
		TypeHints:         make(map[int]string),
	}

	// バイトコードを分析
	instructions := fn.Instructions
	if len(instructions) == 0 {
		return nil, fmt.Errorf("empty function")
	}

	// ネイティブコードを生成
	nativeCode, err := c.generateNativeCode(ctx, instructions)
	if err != nil {
		return nil, fmt.Errorf("code generation failed: %w", err)
	}

	compileTime := time.Since(startTime)
	c.compiledFunctions++
	c.totalCompileTime += compileTime

	// NativeCodeオブジェクトを作成
	result := &NativeCode{
		Function:    fn,
		Code:        nativeCode,
		Size:        len(nativeCode),
		EntryPoint:  unsafe.Pointer(&nativeCode[0]),
		CompileTime: compileTime.Nanoseconds(),
		Metadata: &Metadata{
			OriginalSize:     len(instructions),
			OptimizationInfo: fmt.Sprintf("Level %d", optimizationLevel),
			TypeProfile:      make(map[int]string),
		},
	}

	return result, nil
}

// generateNativeCode はバイトコードからネイティブコードを生成
func (c *Compiler) generateNativeCode(ctx *CompilationContext, instructions code.Instructions) ([]byte, error) {
	// 基本的なx86-64ネイティブコード生成
	// 現在は非常にシンプルな実装：関数プロローグ + エピローグのみ

	var codeBuffer []byte

	// 関数プロローグ
	codeBuffer = append(codeBuffer, c.generatePrologue()...)

	// バイトコード命令をネイティブコードに変換
	ip := 0
	for ip < len(instructions) {
		opcode := code.Opcode(instructions[ip])

		switch opcode {
		case code.OpConstant:
			// 定数ロード
			constIndex := int(instructions[ip+1])<<8 | int(instructions[ip+2])
			nativeInst, err := c.generateConstantLoad(constIndex)
			if err != nil {
				return nil, err
			}
			codeBuffer = append(codeBuffer, nativeInst...)
			ip += 3

		case code.OpAdd:
			// 加算
			nativeInst := c.generateAdd()
			codeBuffer = append(codeBuffer, nativeInst...)
			ip++

		case code.OpSub:
			// 減算
			nativeInst := c.generateSub()
			codeBuffer = append(codeBuffer, nativeInst...)
			ip++

		case code.OpMul:
			// 乗算
			nativeInst := c.generateMul()
			codeBuffer = append(codeBuffer, nativeInst...)
			ip++

		case code.OpDiv:
			// 除算
			nativeInst := c.generateDiv()
			codeBuffer = append(codeBuffer, nativeInst...)
			ip++

		case code.OpPop:
			// スタックポップ
			nativeInst := c.generatePop()
			codeBuffer = append(codeBuffer, nativeInst...)
			ip++

		case code.OpReturn, code.OpReturnValue:
			// 関数リターン
			nativeInst := c.generateReturn()
			codeBuffer = append(codeBuffer, nativeInst...)
			ip++

		default:
			// 未対応の命令はVMフォールバックを使用
			nativeInst := c.generateVMFallback(opcode)
			codeBuffer = append(codeBuffer, nativeInst...)
			ip++
		}
	}

	// 関数エピローグ
	codeBuffer = append(codeBuffer, c.generateEpilogue()...)

	return codeBuffer, nil
}

// generatePrologue は関数プロローグを生成（x86-64）
func (c *Compiler) generatePrologue() []byte {
	// push rbp; mov rbp, rsp
	return []byte{
		0x55,             // push rbp
		0x48, 0x89, 0xe5, // mov rbp, rsp
	}
}

// generateEpilogue は関数エピローグを生成（x86-64）
func (c *Compiler) generateEpilogue() []byte {
	// mov rsp, rbp; pop rbp; ret
	return []byte{
		0x48, 0x89, 0xec, // mov rsp, rbp
		0x5d, // pop rbp
		0xc3, // ret
	}
}

// generateConstantLoad は定数ロードのコードを生成
func (c *Compiler) generateConstantLoad(constIndex int) ([]byte, error) {
	// 簡易実装：定数インデックスをレジスタにロード
	// mov rax, constIndex
	return []byte{
		0x48, 0xb8, // mov rax, imm64
		byte(constIndex), 0, 0, 0, 0, 0, 0, 0, // 64-bit immediate
	}, nil
}

// generateAdd は加算のコードを生成
func (c *Compiler) generateAdd() []byte {
	// 簡易実装：レジスタ間加算
	// add rax, rbx
	return []byte{
		0x48, 0x01, 0xd8, // add rax, rbx
	}
}

// generateSub は減算のコードを生成
func (c *Compiler) generateSub() []byte {
	// sub rax, rbx
	return []byte{
		0x48, 0x29, 0xd8, // sub rax, rbx
	}
}

// generateMul は乗算のコードを生成
func (c *Compiler) generateMul() []byte {
	// imul rax, rbx
	return []byte{
		0x48, 0x0f, 0xaf, 0xc3, // imul rax, rbx
	}
}

// generateDiv は除算のコードを生成
func (c *Compiler) generateDiv() []byte {
	// 簡易実装：VMフォールバックを呼び出し
	return c.generateVMFallback(code.OpDiv)
}

// generatePop はスタックポップのコードを生成
func (c *Compiler) generatePop() []byte {
	// pop rax
	return []byte{
		0x58, // pop rax
	}
}

// generateReturn はリターンのコードを生成
func (c *Compiler) generateReturn() []byte {
	// エピローグと同じ
	return c.generateEpilogue()
}

// generateVMFallback はVM実行にフォールバックするコードを生成
func (c *Compiler) generateVMFallback(opcode code.Opcode) []byte {
	// 簡易実装：VMコール命令
	// call vm_execute_instruction
	return []byte{
		0xe8, 0x00, 0x00, 0x00, 0x00, // call rel32 (placeholder)
	}
}

// AnalyzeBytecode はバイトコードを分析して最適化ヒントを取得
func (c *Compiler) AnalyzeBytecode(instructions code.Instructions) map[string]interface{} {
	analysis := make(map[string]interface{})

	// 命令数
	analysis["instruction_count"] = len(instructions)

	// 使用される命令の種類
	opcodeCount := make(map[code.Opcode]int)
	ip := 0
	for ip < len(instructions) {
		if ip >= len(instructions) {
			break
		}

		opcode := code.Opcode(instructions[ip])
		opcodeCount[opcode]++

		// 命令のオペランド数に基づいてIPを進める
		switch opcode {
		case code.OpConstant, code.OpSetGlobal, code.OpGetGlobal:
			ip += 3 // 2-byte operand
		case code.OpCall, code.OpGetLocal, code.OpSetLocal:
			ip += 2 // 1-byte operand
		default:
			ip += 1 // no operand
		}
	}

	analysis["opcode_distribution"] = opcodeCount

	// 最も多用される命令
	var maxOpcode code.Opcode
	var maxCount int
	for opcode, count := range opcodeCount {
		if count > maxCount {
			maxOpcode = opcode
			maxCount = count
		}
	}
	analysis["most_frequent_opcode"] = maxOpcode
	analysis["most_frequent_count"] = maxCount

	return analysis
}

// GetStatistics はコンパイラーの統計情報を取得
func (c *Compiler) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["compiled_functions"] = c.compiledFunctions
	stats["total_compile_time_ms"] = float64(c.totalCompileTime.Nanoseconds()) / 1e6

	if c.compiledFunctions > 0 {
		avgTime := float64(c.totalCompileTime.Nanoseconds()) / float64(c.compiledFunctions) / 1e6
		stats["average_compile_time_ms"] = avgTime
	}

	return stats
}
