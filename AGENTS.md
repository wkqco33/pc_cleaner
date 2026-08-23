# AGENTS.md — 개발 가이드

이 문서는 PC Cleaner(`pc_cleaner`) 저장소에서 작업하는 모든 에이전트가 반드시 따라야 할 규칙입니다.
**이 프로젝트는 TDD(Test-Driven Development)로 개발됩니다.** 기능 구현 시 테스트를 먼저 작성하고 그 다음 코드를 작성하십시오.

---

## 1. 프로젝트 개요

- **역할**: macOS / Windows / Linux에서 캐시 및 임시 파일을 정리해 디스크 공간을 확보하는 CLI 도구.
- **언어**: Go (`go 1.26.1`)
- **CLI 프레임워크**: [`github.com/wkqco33/wcli`](https://github.com/wkqco33/wcli) (v0.2.0) — 명령/플래그/헬프/버전/셀 완성 제공. 터미널 렌더링은 `wcli/rich` 사용.
- **진입점**: `main.go` (package `main`)
- **패키지 구조** (코드가 `internal/` 아래에 계층화되어 있음):

  ```text
  main.go                       # CLI 진입점: wcli.Command 정의 + 흐름 조합 (thin layer)
  internal/
    cleaner/                    # 실제 삭제 로직 (dir/glob/command) — rich로 출력
    scanner/                    # 캐시 경로 발견 + 용량 계산, OS별 항목
    ui/                         # FormatBytes 바이트 포맷터만 제공 (색상/렌더링은 rich)
  ```

- **빌드/실행 도구**: `Taskfile.yml` (`task build`, `task test`, `task run`, `task clean`).

---

## 2. TDD 워크플로우 (필수)

**모든 기능/버그수정은 다음 사이클을 반드시 따른다. 순서를 바꾸지 말 것.**

1. **RED** — 먼저 실패하는 테스트를 작성한다.
   - 검증하고 싶은 동작/시나리오를 테스트로 정의한다.
   - 이 시점에 `go test ./...` 는 반드시 **실패**(컴파일 오류 포함)해야 한다.
2. **GREEN** — 그 테스트를 통과시키는 **최소한의** 코드만 작성한다.
   - `go test ./...` 가 통과할 때까지.
3. **REFACTOR** — 중복 제거, 네이밍, 구조 정리. 이때도 테스트는 계속 통과해야 한다.

> ⚠️ 구현이 먼저이고 테스트가 나중이면 안 된다. 기능 구현을 시작하기 전에 반드시 실패 테스트가 존재해야 한다.
> 기존 코드에 테스트가 없는 경우라면 "캐릭터화 테스트(characterization test)"로 먼저 현재 동작을 고정한 뒤 변경한다.

### 작업 순서 예시

```bash
# 1) 실패 테스트 작성 후 실행
go test ./...

# 2) 구현 후 실행
go test ./...

# 3) 리팩터링 후 커버리지/품질 확인
go test ./... -cover
go vet ./...
```

---

## 3. 테스트 실행 방법

```bash
go test ./...            # 전체 테스트
go test ./... -v         # 상세 출력
go test ./... -cover     # 커버리지 확인
go test ./<pkg> -run <TestName>   # 특정 테스트만
go vet ./...             # 정적 검사
task test                # Taskfile 경유 (동일)
task coverage            # 커버리지 게이트 (핵심 패키지 >= 70%%)
```

CI(GitHub Actions `.github/workflows/ci.yml`)에서도 `go test ./... -v`와 **핵심 패키지(`scanner`, `cleaner`) 커버리지 70% 이상 게이트**가 실행됩니다. 커버리지가 70% 미만이면 CI가 실패하므로, 변경 후 반드시 `task coverage`로 확인하십시오.

워크플로 분리 규칙:

- `ci.yml` — push/PR 시 **린트(gofmt·go vet)·테스트·커버리지·크로스컴파일**을 실행.
- `release.yml` — **`v*` 태그 push 시에만** 아카이브+체크섬 빌드 후 GitHub 릴리스를 생성.
- 새 태그 배포: `git tag v1.2.3 && git push origin v1.2.3`

---

## 4. 테스트 작성 규칙

### 4.1 파일 위치 / 패키지

- 테스트 파일은 테스트할 코드와 **같은 디렉토리**에 `*_test.go`로 둔다.
- 패키지 내부(단위) 테스트가 필요한 경우 `package <pkg>` (내부 테스트)를 쓰고, 공개 API만 검증할 때는 `package <pkg>_test` (외부/블랙박스 테스트)를 쓴다.
  - 예: `internal/ui/ui_output_test.go`는 `package ui` (내부) — 비공개 `output`/`colorEnabled` 접근 필요.

### 4.2 네이밍

- 테스트 함수는 `Test<Subject>_<Scenario>` 형태를 사용한다.
  - 예: `TestClean_DryRun`, `TestScanItem_NotExist`, `TestFilterItems_IsCaseInsensitive`.
- 검증 헬퍼는 `t.Helper()`를 첫 줄에 호출하고 `t.Fatal`/`t.Errorf`를 사용한다.

### 4.3 필수 관례

- **임시 자원**: 실제 파일시스템은 절대 쓰지 말고 `t.TempDir()` 사용 (자동 정리). 실제 시스템 캐시 경로를 지우면 안 된다.
- **플랫폼 의존 테스트**: OS가 필요한 경우 `t.Skip`으로 처리 (예: `TestGetDiskUsage`의 홈 디렉토리).
- **Table-driven**: 같은 입력 형태가 여러 개면 테이블 드라이브 테스트 사용 (예: `TestFormatBytes`).
- **출력 검증**: `ui` 패키지는 주입 가능한 `output io.Writer`를 제공한다. 출력 테스트는 이를 `bytes.Buffer`로 교체하고 `colorEnabled = false`로 ANSI를 끄고 검증한다. (`internal/ui/ui_output_test.go` 참고)
- **병렬 스캔**: `scanner.Scan`은 고루틴을 쓰므로 결과 개수/합이 정확한지 검증.

### 4.4 커버리지 기준

- 핵심 로직 패키지(`scanner`, `cleaner`)는 **70% 이상** 유지한다.
- 신규 로직은 커버리지를 낮추는 것보다 검증이 우선이다. 가능하면 새 기능에는 테스트를 함께 추가한다.

---

## 5. 아키텍처/설계 제약

### 5.1 내부 패키지 수직 구조

- `main.go`는 **thin layer**로만 유지한다. 비즈니스 로직은 `internal/` 패키지로 내려 보낸다.
- 패키지 간 의존성: `main` → `cleaner` → `scanner`, `ui`, 그리고 외부 `wcli`(+`wcli/rich`). `cleaner`가 `scanner`를 의존하는 것은 허용. 역방향 의존(아래 계층이 위 계층 참조)은 금지.

### 5.2 CLI 작성 규칙 (wcli)

- CLI는 반드시 `wcli.Command` 기반으로 구성한다. stdlib `flag`를 직접 쓰지 않는다.
- 출력은 `rich.Fprintln(w, "[tag]...[/tag]", ...)` 또는 `rich.NewTable(...)`을 사용한다.
- **마크업 태그는 rich의 유효 태그만 사용**한다(`bold`/`dim`/`red`/`green`/`yellow`/`blue`/`magenta`/`cyan`/`white`/`bg-*` 등). `[gray]`처럼 없는 태그는 문자 그대로 출력되므로 사용 금지. 회색 계열은 `[dim]`을 쓴다.
- 테스트/출력 결정성: `rich.Fprint/Fprintln`은 non-terminal writer(`bytes.Buffer` 등)에선 ANSI를 자동 제거하므로, 출력 검증은 `newApp(&buf, ...)` 식으로 버퍼를 주입해 검증한다. (`main_test.go` 참고)
- `--version`은 `wcli.Command.Version` 필드로 자동 등록한다. 셸 완성은 `wcli.NewCompletionCommand(root)`를 `AddCommand`로 붙인다.

### 5.3 OS별 코드

- `internal/scanner/`의 OS별 항목 함수는 build tag **없이** 모든 플랫폼에서 컴파일된다 → 어느 플랫폼에서든 테스트 가능.
  - `commonItems()`, `darwinItems()`, `linuxItems()`, `windowsItems()`.
- 플랫폼 전용 구현(예: `isElevated`, `getPlatformDiskUsage`)은 반드시 build tag(`//go:build windows` / `//go:build !windows` 등)로 분리한다.
- 새 캐시 항목을 추가할 때는 반드시 불변식 테스트(`verifyItemList`)를 통과시켜야 한다: `Name`/`Category` 비어있지 않음, `Type` 유효, `Dir/Glob`은 `Path`, `Command`는 `Command` 필수, 중첩 디렉토리 dedup.

### 5.3 테스트 용이성

- 전역 가변 상태(출력 writer, 색상 플래그)는 패키지 변수로 두되 테스트에서 교체할 수 있게 주입 지점을 마련한다.
- `exec.Command`, `os.*` 호출은 그대로 둘 수 있으나, 그 로직은 실제 동작이 아닌 격리된 임시 자원으로 테스트한다.

---

## 6. 완료 전 체크리스트

브랜치를 완료하기 전에 반드시 아래를 모두 만족시켜야 한다.

```bash
go test ./... -v       # 모든 테스트 통과
go test ./... -cover   # 커버리지 기준 충족
task coverage          # 커버리지 게이트 (핵심 패키지 >= 70%%)
go vet ./...           # 에러 없음
gofmt -l .             # (아무 파일도 출력되지 않아야 함 — gofmt 정렬 완료)
go build ./...         # 컴파일 성공
```

- 기능 추가/버그 수정 시 **실패 테스트 → 구현 → 리팩터** 순서를 지켰는지 확인.
- 테스트 없이 코드만 추가하는 PR은 거절한다.
- OS별 동작을 바꾸는 경우에는 해당 OS 관련 테스트를 추가/수정한다.

---

## 7. 협업 규칙

- 변경 작업을 시작하기 전에 테스트가 실패하는 상태(RED)를 먼저 만들고 그 단계를 기록한다.
- 각 커밋은 "RED → GREEN → REFACTOR" 사이클을 단위로 하면 좋다.
- 라이선스: MIT. 명령어 실행(`docker system prune`, `journalctl` 등)은 실사용 시 주의해야 하므로 테스트는 `whoami` 같은 무해한 명령으로만 검증한다.

---

> 이 문서의 목적은 **에이전트가 코드만 찍어내는 것이 아니라, 검증 가능한 동작을 먼저 정의하고 구현하도록 강제**하는 것입니다. TDD를 어기면 코드가 좁더러진채로 남습니다.
