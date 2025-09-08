package jit

import (
	"monkey/object"
	"sync"
	"time"
)

// FunctionProfile は関数のプロファイル情報
type FunctionProfile struct {
	Function      *object.CompiledFunction // 関数への参照
	CallCount     int64                    // 呼び出し回数
	LastCallTime  time.Time                // 最後の呼び出し時間
	TotalTime     time.Duration            // 累積実行時間
	ArgumentTypes map[int]map[string]int64 // 引数の型情報 [arg_index][type_name]count
	ReturnTypes   map[string]int64         // 戻り値の型情報
}

// Profiler はJITコンパイルのためのプロファイラー
type Profiler struct {
	// 関数プロファイル
	profiles map[*object.CompiledFunction]*FunctionProfile
	mutex    sync.RWMutex

	// 設定
	enabled bool

	// 統計
	totalCalls int64
}

// NewProfiler は新しいプロファイラーを作成
func NewProfiler() *Profiler {
	return &Profiler{
		profiles: make(map[*object.CompiledFunction]*FunctionProfile),
		enabled:  true,
	}
}

// SetEnabled はプロファイリングの有効/無効を設定
func (p *Profiler) SetEnabled(enabled bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.enabled = enabled
}

// RecordCall は関数呼び出しを記録
func (p *Profiler) RecordCall(fn *object.CompiledFunction, args []object.Object) {
	if !p.enabled {
		return
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// プロファイル取得または作成
	profile, exists := p.profiles[fn]
	if !exists {
		profile = &FunctionProfile{
			Function:      fn,
			CallCount:     0,
			ArgumentTypes: make(map[int]map[string]int64),
			ReturnTypes:   make(map[string]int64),
		}
		p.profiles[fn] = profile
	}

	// 呼び出し情報更新
	profile.CallCount++
	profile.LastCallTime = time.Now()
	p.totalCalls++

	// 引数の型情報を記録
	for i, arg := range args {
		if profile.ArgumentTypes[i] == nil {
			profile.ArgumentTypes[i] = make(map[string]int64)
		}
		typeName := string(arg.Type())
		profile.ArgumentTypes[i][typeName]++
	}
}

// RecordReturn は関数の戻り値を記録
func (p *Profiler) RecordReturn(fn *object.CompiledFunction, result object.Object) {
	if !p.enabled {
		return
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	profile, exists := p.profiles[fn]
	if !exists {
		return
	}

	// 戻り値の型情報を記録
	if result != nil {
		typeName := string(result.Type())
		profile.ReturnTypes[typeName]++
	}
}

// RecordExecutionTime は実行時間を記録
func (p *Profiler) RecordExecutionTime(fn *object.CompiledFunction, duration time.Duration) {
	if !p.enabled {
		return
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	profile, exists := p.profiles[fn]
	if exists {
		profile.TotalTime += duration
	}
}

// ShouldCompile はJITコンパイルを実行すべきかを判定
func (p *Profiler) ShouldCompile(fn *object.CompiledFunction, threshold int) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	profile, exists := p.profiles[fn]
	if !exists {
		return false
	}

	return profile.CallCount >= int64(threshold)
}

// GetProfile は関数のプロファイル情報を取得
func (p *Profiler) GetProfile(fn *object.CompiledFunction) (*FunctionProfile, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	profile, exists := p.profiles[fn]
	if !exists {
		return nil, false
	}

	// プロファイルのコピーを返す（安全性のため）
	copy := &FunctionProfile{
		Function:      profile.Function,
		CallCount:     profile.CallCount,
		LastCallTime:  profile.LastCallTime,
		TotalTime:     profile.TotalTime,
		ArgumentTypes: make(map[int]map[string]int64),
		ReturnTypes:   make(map[string]int64),
	}

	// 引数型情報をコピー
	for argIdx, types := range profile.ArgumentTypes {
		copy.ArgumentTypes[argIdx] = make(map[string]int64)
		for typeName, count := range types {
			copy.ArgumentTypes[argIdx][typeName] = count
		}
	}

	// 戻り値型情報をコピー
	for typeName, count := range profile.ReturnTypes {
		copy.ReturnTypes[typeName] = count
	}

	return copy, true
}

// GetHotFunctions は最もよく呼ばれる関数のリストを取得
func (p *Profiler) GetHotFunctions(limit int) []*FunctionProfile {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// プロファイルをスライスに変換
	profiles := make([]*FunctionProfile, 0, len(p.profiles))
	for _, profile := range p.profiles {
		profiles = append(profiles, profile)
	}

	// 呼び出し回数でソート（バブルソート - シンプルな実装）
	for i := 0; i < len(profiles)-1; i++ {
		for j := 0; j < len(profiles)-i-1; j++ {
			if profiles[j].CallCount < profiles[j+1].CallCount {
				profiles[j], profiles[j+1] = profiles[j+1], profiles[j]
			}
		}
	}

	// 上位limit個を返す
	if limit > len(profiles) {
		limit = len(profiles)
	}

	result := make([]*FunctionProfile, limit)
	for i := 0; i < limit; i++ {
		// プロファイルのコピーを作成
		original := profiles[i]
		result[i] = &FunctionProfile{
			Function:      original.Function,
			CallCount:     original.CallCount,
			LastCallTime:  original.LastCallTime,
			TotalTime:     original.TotalTime,
			ArgumentTypes: make(map[int]map[string]int64),
			ReturnTypes:   make(map[string]int64),
		}

		// 型情報をコピー
		for argIdx, types := range original.ArgumentTypes {
			result[i].ArgumentTypes[argIdx] = make(map[string]int64)
			for typeName, count := range types {
				result[i].ArgumentTypes[argIdx][typeName] = count
			}
		}

		for typeName, count := range original.ReturnTypes {
			result[i].ReturnTypes[typeName] = count
		}
	}

	return result
}

// GetDominantArgumentType は最も頻繁に使われる引数の型を取得
func (p *Profiler) GetDominantArgumentType(fn *object.CompiledFunction, argIndex int) (string, int64) {
	profile, exists := p.GetProfile(fn)
	if !exists {
		return "", 0
	}

	types, exists := profile.ArgumentTypes[argIndex]
	if !exists {
		return "", 0
	}

	var maxType string
	var maxCount int64

	for typeName, count := range types {
		if count > maxCount {
			maxType = typeName
			maxCount = count
		}
	}

	return maxType, maxCount
}

// GetDominantReturnType は最も頻繁に返される型を取得
func (p *Profiler) GetDominantReturnType(fn *object.CompiledFunction) (string, int64) {
	profile, exists := p.GetProfile(fn)
	if !exists {
		return "", 0
	}

	var maxType string
	var maxCount int64

	for typeName, count := range profile.ReturnTypes {
		if count > maxCount {
			maxType = typeName
			maxCount = count
		}
	}

	return maxType, maxCount
}

// GetTotalCalls は総呼び出し回数を取得
func (p *Profiler) GetTotalCalls() int64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.totalCalls
}

// Reset はプロファイル情報をリセット
func (p *Profiler) Reset() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.profiles = make(map[*object.CompiledFunction]*FunctionProfile)
	p.totalCalls = 0
}

// GetProfileCount はプロファイルされた関数の数を取得
func (p *Profiler) GetProfileCount() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.profiles)
}
