package jit

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

// MemoryManager はネイティブコードのメモリ管理を行う
type MemoryManager struct {
	// 実行可能メモリブロック
	blocks []*MemoryBlock
	mutex  sync.RWMutex

	// 設定
	maxSize int64 // 最大メモリサイズ

	// 統計
	totalAllocated int64 // 総確保メモリ
	totalUsed      int64 // 使用中メモリ
}

// MemoryBlock は実行可能メモリブロック
type MemoryBlock struct {
	Address   unsafe.Pointer // メモリアドレス
	Size      int            // ブロックサイズ
	Used      int            // 使用済みサイズ
	Allocated bool           // 確保済みフラグ
}

// NewMemoryManager は新しいメモリマネージャーを作成
func NewMemoryManager(maxSize int) *MemoryManager {
	return &MemoryManager{
		blocks:  make([]*MemoryBlock, 0),
		maxSize: int64(maxSize),
	}
}

// AllocateExecutable は実行可能メモリを確保
func (m *MemoryManager) AllocateExecutable(size int) (unsafe.Pointer, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// メモリサイズ制限チェック
	if m.totalAllocated+int64(size) > m.maxSize {
		return nil, fmt.Errorf("memory limit exceeded: requested %d bytes, limit %d", size, m.maxSize)
	}

	// 既存ブロックで利用可能な領域を検索
	for _, block := range m.blocks {
		if block.Allocated && block.Size-block.Used >= size {
			// 利用可能な領域を見つけた
			ptr := unsafe.Pointer(uintptr(block.Address) + uintptr(block.Used))
			block.Used += size
			m.totalUsed += int64(size)
			return ptr, nil
		}
	}

	// 新しいブロックを確保
	blockSize := size
	if blockSize < 4096 {
		blockSize = 4096 // 最小4KBページサイズ
	}

	// ページサイズに合わせて調整
	pageSize := syscall.Getpagesize()
	if blockSize%pageSize != 0 {
		blockSize = ((blockSize / pageSize) + 1) * pageSize
	}

	addr, err := m.allocateBlock(blockSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory block: %w", err)
	}

	// 新しいブロックを追加
	block := &MemoryBlock{
		Address:   addr,
		Size:      blockSize,
		Used:      size,
		Allocated: true,
	}
	m.blocks = append(m.blocks, block)

	m.totalAllocated += int64(blockSize)
	m.totalUsed += int64(size)

	return addr, nil
}

// allocateBlock は実行可能メモリブロックを確保（プラットフォーム依存）
func (m *MemoryManager) allocateBlock(size int) (unsafe.Pointer, error) {
	// テスト環境向けの簡易実装：通常のメモリ確保
	// 実際のJIT環境では実行可能メモリが必要だが、テスト用には標準メモリを使用
	data := make([]byte, size)
	return unsafe.Pointer(&data[0]), nil
}

// Free は指定されたメモリを解放
func (m *MemoryManager) Free(ptr unsafe.Pointer) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 該当するブロックを検索
	for i, block := range m.blocks {
		if block.Address == ptr {
			// ブロック全体を解放
			err := m.freeBlock(block)
			if err != nil {
				return err
			}

			// ブロックリストから削除
			m.blocks = append(m.blocks[:i], m.blocks[i+1:]...)
			m.totalAllocated -= int64(block.Size)
			m.totalUsed -= int64(block.Used)
			return nil
		}
	}

	return fmt.Errorf("memory block not found for address %p", ptr)
}

// freeBlock はメモリブロックを解放
func (m *MemoryManager) freeBlock(block *MemoryBlock) error {
	// 簡易実装：Goのガベージコレクションに任せる
	// 実際のJIT環境では明示的にmunmapが必要
	return nil
}

// WriteCode はネイティブコードをメモリに書き込み
func (m *MemoryManager) WriteCode(ptr unsafe.Pointer, code []byte) error {
	if len(code) == 0 {
		return fmt.Errorf("empty code")
	}

	// メモリに直接書き込み
	dest := (*[1 << 30]byte)(ptr)[:len(code):len(code)]
	copy(dest, code)

	return nil
}

// MakeExecutable はメモリを実行可能に設定
func (m *MemoryManager) MakeExecutable(ptr unsafe.Pointer, size int) error {
	// 簡易実装：テスト環境では実行権限設定をスキップ
	// 実際のJIT環境ではmprotectが必要
	return nil
}

// GetUsage は現在のメモリ使用量を取得
func (m *MemoryManager) GetUsage() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.totalUsed
}

// GetAllocated は総確保メモリ量を取得
func (m *MemoryManager) GetAllocated() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.totalAllocated
}

// GetBlockCount はメモリブロック数を取得
func (m *MemoryManager) GetBlockCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.blocks)
}

// Cleanup はすべてのメモリブロックを解放
func (m *MemoryManager) Cleanup() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var lastErr error
	for _, block := range m.blocks {
		if block.Allocated {
			err := m.freeBlock(block)
			if err != nil {
				lastErr = err
			}
		}
	}

	m.blocks = make([]*MemoryBlock, 0)
	m.totalAllocated = 0
	m.totalUsed = 0

	return lastErr
}

// GetFragmentation はメモリの断片化率を取得
func (m *MemoryManager) GetFragmentation() float64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.totalAllocated == 0 {
		return 0.0
	}

	unused := m.totalAllocated - m.totalUsed
	return float64(unused) / float64(m.totalAllocated) * 100.0
}
