# PC Cleaner

Mac / Windows / Linux에서 캐시 및 임시 파일을 정리해 디스크 공간을 확보하는 CLI 도구.

## 설치

### 바이너리 다운로드

릴리즈 페이지에서 OS에 맞는 바이너리를 다운로드하거나, 직접 빌드합니다.

### 직접 빌드

```bash
git clone https://github.com/seomini/pc_cleaner
cd pc_cleaner
go build -o pc_cleaner .
```

## 사용법

```bash
# 분석만 실행 (삭제 없음)
./pc_cleaner --dry-run

# 실행 (스캔 후 확인 요청)
./pc_cleaner

# 특정 항목 제외
./pc_cleaner --skip=gradle,docker

# 버전 확인
./pc_cleaner --version
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
GOOS=linux  GOARCH=amd64 go build -o bin/pc_cleaner_linux_amd64 .
GOOS=windows GOARCH=amd64 go build -o bin/pc_cleaner_windows_amd64.exe .
GOOS=darwin  GOARCH=arm64 go build -o bin/pc_cleaner_darwin_arm64 .
```

## 라이센스

[MIT License](./LICENSE)
