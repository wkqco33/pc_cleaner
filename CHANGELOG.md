# Changelog

모든 주요 변경 사항을 이 문서에 기록합니다.

## [Unreleased]

### Added

- 비대화형 실행을 위한 `--yes`, `--no-input`, `--quiet`, `--no-color` 플래그
- `--format json` 기계 판독 출력
- 위험 경로 차단과 설치되지 않은 명령 자동 제외
- 자동화 환경의 종료 코드 문서화
- AI 호출 재시도, 지수 백오프, request ID

### Changed

- 정리 진행 출력을 주입된 writer로 전달
- 결과는 stdout, 진행·오류 메시지는 stderr로 분리
- Docker와 journald 명령 인자를 허용된 형태로 제한

## [0.3.0] - 2026-09-12

### Added

- LLM 기반 스마트 정리 기능
