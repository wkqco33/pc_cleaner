# PC Cleaner

Mac / Windows / Linux에서 캐시 및 임시 파일을 정리해 디스크 공간을 확보하는 CLI 도구.

## 설치

### ppm (Private Package Manager, 권장)

[`ppm`](https://github.com/wkqco33/package_manager)이 설치되어 있다면 OS/아키텍처에 맞는 아카이브를 자동으로 내려받아 설치합니다.

```bash
ppm install wkqco33/pc_cleaner
# 계획만 확인: ppm install --dry-run wkqco33/pc_cleaner
```

### go install (권장)

Go가 설치되어 있다면 `go install`로 바로 설치할 수 있습니다.

```bash
go install github.com/wkqco33/pc_cleaner@latest
```

### 바이너리 다운로드

릴리즈 페이지에서 OS에 맞는 바이너리를 다운로드하거나, 직접 빌드합니다.

### 직접 빌드

```bash
git clone https://github.com/wkqco33/pc_cleaner
cd pc_cleaner
go build -o pcc .
```

## 사용법

```bash
# 분석만 실행 (삭제 없음)
./pcc --dry-run

# 실행 (스캔 후 확인 요청)
./pcc

# 특정 항목 제외
./pcc --skip=gradle,docker

# 버전 확인
./pcc --version
```

## AI 스마트 정리 (`pcc ai`)

LLM(기본값: Ollama)을 활용하여 캐시 항목들의 위험도를 평가하고 최적의 정리 계획을 리포트 및 실행합니다.

```bash
# 기본 AI 분석 및 스마트 정리 (기본 Ollama llama3.2:latest)
./pcc ai

# 자연어 지시사항과 함께 실행
./pcc ai "개발 환경 캐시는 건드리지 말고 브라우저와 시스템 임시 파일만 정리해줘"

# AI 분석만 확인 (삭제 없음)
./pcc ai --dry-run

# 질문 없이 추천 항목 즉시 정리
./pcc ai -y

# 커스텀 모델 / 엔드포인트 / 프로바이더 지정
./pcc ai --model=qwen2.5:7b --endpoint=http://localhost:11434/v1
./pcc ai --provider=openai --model=gpt-4o-mini
```


## 정리 대상

### macOS

- `~/Library/Caches` — 사용자 캐시
- `~/Library/Logs` — 앱 로그
- `~/Library/Developer/Xcode/DerivedData` — Xcode 빌드 캐시
- `~/Library/Developer/CoreSimulator/Caches` — iOS 시뮬레이터 캐시
- `~/.Trash` — 휴지통

### Windows

- `%TEMP%`, `C:\Windows\Temp` — 임시 파일
- `C:\Windows\Prefetch` — 프리패치 캐시
- IE/Edge 캐시, 썸네일 캐시

### Linux

- `~/.cache/thumbnails` — 썸네일 캐시
- `/tmp`, `/var/tmp` — 임시 파일
- `journalctl --vacuum-size` — 저널 로그 정리

### 공통 (개발 도구)

| 항목 | 경로 |
| ---- | ---- |
| Gradle | `~/.gradle/caches` |
| Maven | `~/.m2/repository` |
| pip | `~/.cache/pip` |
| uv | `~/.cache/uv` |
| npm | `~/.npm/_cacache` |
| yarn | `~/.yarn/cache` |
| Cargo | `~/.cargo/registry/cache` |
| Go | `~/go/pkg/mod/cache` |
| Docker | `docker system prune -f` |

## 안전 정책

- 존재하지 않는 경로는 자동으로 skip
- 디렉토리 자체는 유지하고 **내용만** 삭제
- 접근 권한이 없는 파일은 skip
- `--dry-run`으로 삭제 없이 미리 확인 가능

## 크로스컴파일

```bash
GOOS=linux  GOARCH=amd64 go build -o bin/pcc_linux_amd64 .
GOOS=windows GOARCH=amd64 go build -o bin/pcc_windows_amd64.exe .
GOOS=darwin  GOARCH=arm64 go build -o bin/pcc_darwin_arm64 .
```

## 라이센스

[MIT License](./LICENSE)
