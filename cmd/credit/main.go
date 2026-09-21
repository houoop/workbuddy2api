// credit.go — WorkBuddy 积分查询（全部账号 + 总计），JSON 输出到 stdout。
//
// 用法:
//
//	go run ./cmd/credit        # 或编译后 ./credit
//
// 输出结构:
//
//	{"service":"workbuddy","ts":N,
//	 "total":{"remain":N,"used":N,"size":N,"accounts":N,"ok":N,"failed":N},
//	 "accounts":[{"uid","nickname","remain","used","size","packages","ok","error?"}]}
//
// realm 感知：复用 upstream.Client（auth.Parse + upstream.New），global 账号查积分
// 走 workbuddy.ai /billing/meter/*（404 回落 /v2），CN 账号维持 codebuddy.cn
// /v2/billing/meter/get-user-resource（现状逐字）。聚合口径即 upstream.ResourceSummary。
package main

import (
	"encoding/json"
	"fmt"
	"os"
<<<<<<< HEAD
	"path/filepath"
	"sort"
	"strings"
=======
>>>>>>> upstream-v2
	"time"

	"workbuddy2api/internal/auth"
	"workbuddy2api/internal/upstream"
)

<<<<<<< HEAD
const billingBaseCN = "https://www.codebuddy.cn"
const billingBaseIntl = "https://www.codebuddy.ai"

type authFile struct {
	Auth struct {
		AccessToken string `json:"accessToken"`
		Domain      string `json:"domain"`
	} `json:"auth"`
	Account struct {
		UID          string `json:"uid"`
		EnterpriseID string `json:"enterpriseId"`
		Nickname     string `json:"nickname"`
	} `json:"account"`
}

=======
>>>>>>> upstream-v2
type accountResult struct {
	UID      string `json:"uid"`
	Nickname string `json:"nickname"`
	Remain   *int64 `json:"remain"`
	Used     *int64 `json:"used"`
	Size     *int64 `json:"size"`
	Packages int    `json:"packages,omitempty"`
	OK       bool   `json:"ok"`
	Error    string `json:"error,omitempty"`
}

<<<<<<< HEAD
type resourcePackage struct {
	CapacityRemain      int64 `json:"CapacityRemain"`
	CapacityUsed        int64 `json:"CapacityUsed"`
	CapacitySize        int64 `json:"CapacitySize"`
	CycleCapacityRemain int64 `json:"CycleCapacityRemain"`
	CycleCapacityUsed   int64 `json:"CycleCapacityUsed"`
	CycleCapacitySize   int64 `json:"CycleCapacitySize"`
}

// packageRemainUsed 与 billing.go:203-258 一致
func packageRemainUsed(a resourcePackage) (remain, used, size int64) {
	if a.CycleCapacitySize > 0 {
		remain = a.CycleCapacityRemain
		size = a.CycleCapacitySize
		if remain < 0 {
			remain = 0
		}
		if remain > size {
			remain = size
		}
		used = size - remain
		if a.CycleCapacityUsed > used {
			used = a.CycleCapacityUsed
			if size >= used {
				remain = size - used
			}
		}
		return remain, used, size
	}
	remain = a.CapacityRemain
	used = a.CapacityUsed
	size = a.CapacitySize
	if used == 0 && size > remain {
		used = size - remain
	}
	return remain, used, size
}

// billingBaseFor 按账号 domain 选择国内/国际版计费域名。
func billingBaseFor(af *authFile) string {
	if strings.Contains(strings.ToLower(af.Auth.Domain), "codebuddy.ai") {
		return billingBaseIntl
	}
	return billingBaseCN
}

func fetchUserResource(af *authFile) (remain, used, size int64, packs int, err error) {
	now := time.Now()
	body, _ := json.Marshal(map[string]any{
		"PageNumber":               1,
		"PageSize":                 100,
		"ProductCode":              "p_tcaca",
		"Status":                   []int{0, 3},
		"PackageEndTimeRangeBegin": now.Format("2006-01-02 15:04:05"),
		"PackageEndTimeRangeEnd":   now.Add(365 * 101 * 24 * time.Hour).Format("2006-01-02 15:04:05"),
	})
	req, err := http.NewRequest(http.MethodPost, billingBaseFor(af)+"/v2/billing/meter/get-user-resource", bytes.NewReader(body))
	if err != nil {
		return 0, 0, 0, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+af.Auth.AccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if af.Account.UID != "" {
		req.Header.Set("X-User-Id", af.Account.UID)
	}
	if af.Account.EnterpriseID != "" {
		req.Header.Set("X-Enterprise-Id", af.Account.EnterpriseID)
		req.Header.Set("X-Tenant-Id", af.Account.EnterpriseID)
	}
	if af.Auth.Domain != "" {
		req.Header.Set("X-Domain", af.Auth.Domain)
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return 0, 0, 0, 0, fmt.Errorf("http %d", resp.StatusCode)
	}
	var env struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Response struct {
				Data struct {
					TotalDosage int64             `json:"TotalDosage"`
					Accounts    []resourcePackage `json:"Accounts"`
				} `json:"Data"`
			} `json:"Response"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return 0, 0, 0, 0, err
	}
	if env.Code != 0 {
		return 0, 0, 0, 0, fmt.Errorf("code=%d %s", env.Code, env.Msg)
	}
	for _, a := range env.Data.Response.Data.Accounts {
		r, u, s := packageRemainUsed(a)
		remain += r
		used += u
		size += s
	}
	packs = len(env.Data.Response.Data.Accounts)
	if size > 0 {
		if derived := size - remain; derived > used {
			used = derived
		}
	}
	if dosage := env.Data.Response.Data.TotalDosage; dosage > size {
		size = dosage
		if derived := size - remain; derived > used {
			used = derived
		}
	}
	return remain, used, size, packs, nil
}

=======
>>>>>>> upstream-v2
func main() {
	pretty := len(os.Args) > 1 && os.Args[1] == "-pretty"
	authDir := "./auths"
	if v := os.Getenv("WB2A_AUTH_DIR"); v != "" {
		authDir = v
	}
	up := upstream.New()
	up.GlobalEnabled = true // 允许按 realm 路由：global 账查积分走 workbuddy.ai
	accounts := collect(authDir, up)
	printAccounts(accounts, pretty)
}

// collect 遍历 auths 目录并查询每个账号的积分摘要。供测试注入 fake upstream 断言
// realm 路由（main 从 os.Args/env 取况，collect 单一来源可测）。
// 文件清单走 auth.LoadAuthFiles（宽侧 workbuddy*.json）：与网关 LoadDir 同口径，
// 不带连字符的文件不再被跳过（P2-10，审查发现 10）。
func collect(authDir string, up *upstream.Client) []accountResult {
	files, _ := auth.LoadAuthFiles(authDir)
	accounts := make([]accountResult, 0, len(files))
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		a, err := auth.Parse(raw)
		if err != nil {
			continue
		}
		res := accountResult{UID: a.UID, Nickname: a.Nickname}
		if a.AccessToken == "" {
			res.Error = "no accessToken"
			accounts = append(accounts, res)
			continue
		}
		remain, used, size, packs, err := up.ResourceSummary(a)
		if err != nil {
			res.Error = err.Error()
		} else {
			res.Remain = &remain
			res.Used = &used
			res.Size = &size
			res.Packages = packs
			res.OK = true
		}
		accounts = append(accounts, res)
		time.Sleep(200 * time.Millisecond)
	}
	return accounts
}

// printAccounts 汇总并输出结果：-pretty 走人类可读日报，否则 JSON（与老版输出一致）。
func printAccounts(accounts []accountResult, pretty bool) {
	var totalRemain, totalUsed, totalSize int64
	okCount := 0
	for _, a := range accounts {
		if a.OK {
			okCount++
			if a.Remain != nil {
				totalRemain += *a.Remain
			}
			if a.Used != nil {
				totalUsed += *a.Used
			}
			if a.Size != nil {
				totalSize += *a.Size
			}
		}
	}
	if pretty {
		printPretty(accounts, totalRemain, totalUsed, totalSize, okCount)
		return
	}
	out := map[string]any{
		"service": "workbuddy",
		"ts":      time.Now().Unix(),
		"total": map[string]any{
			"remain":   totalRemain,
			"used":     totalUsed,
			"size":     totalSize,
			"accounts": len(accounts),
			"ok":       okCount,
			"failed":   len(accounts) - okCount,
		},
		"accounts": accounts,
	}
	raw, _ := json.Marshal(out)
	fmt.Println(string(raw))
}

// printPretty 人类可读日报：四行汇总，无账号明细。
func printPretty(accounts []accountResult, totalRemain, totalUsed, totalSize int64, okCount int) {
	withBalance := 0
	var failed []string
	for _, a := range accounts {
		if a.OK && a.Remain != nil && *a.Remain > 0 {
			withBalance++
		}
		if !a.OK {
			name := a.Nickname
			if name == "" && len(a.UID) >= 8 {
				name = a.UID[:8]
			}
			failed = append(failed, name+" "+a.Error)
		}
	}
	pct := int64(0)
	if totalSize > 0 {
		pct = totalRemain * 100 / totalSize
	}
	fmt.Printf("📊 WorkBuddy 积分日报\n")
	fmt.Printf("账号: %d/%d\n", withBalance, len(accounts))
	fmt.Printf("总计: %d/%d\n", totalRemain, totalSize)
	fmt.Printf("剩余: %d%%\n", pct)
	for _, f := range failed {
		fmt.Printf("⚠️ %s\n", f)
	}
}