# AGENTS.md

## 목적
새 세션 에이전트가 **빠르게 안전하게 이어서 작업**할 수 있도록, 코드에서 바로 드러나지 않는 의사결정/운영 규칙만 기록한다.

## 하드 규칙 (중요)
- exe 산출물은 항상 `out\game-save-backup-manager.exe`.
- 아이콘 포함 exe 빌드는 `go build`가 아니라 `fyne package --icon assets\app-icon.png` 경로를 사용.
- 메인 창 닫힘 시 `DB 관리` 창도 함께 닫히도록 이미 구현되어 있음(창 라이프사이클 연동).
- `reg` 경로 해석 규칙은 사용자 합의로 고정:
- 경로 끝이 `\\`면 키 전체, 아니면 값 단일.
- 이 규칙은 스캔/백업/검증 모두 동일하게 따라야 함.

## 놓치기 쉬운 동작
- `reg`의 `경로 열기`는 파일 탐색기가 아니라 레지스트리 편집기 실행.
- 일반 실행 실패 시 `RunAs`로 UAC 승격 재시도.
- `zip` 백업은 ZIP 내부에 디렉터리 계층을 보존하며, 환경변수 표기(예: `%APPDATA%`) 논리 경로도 유지되도록 처리됨.
- 메인 화면은 전체 DB 항목이 아니라 “감지된 세이브만” 보여주는 리스트 UI.

## 현재 UI 계약
- 메인 상단: 왼쪽 `스캔`, 오른쪽 `DB 관리`.
- 메인 목록 행: 게임명 + `백업` + `경로 열기`.
- `DB 관리` 창: 등록/수정/삭제/새로고침 + 테이블.
- `type` 입력은 드롭다운(`zip`, `reg`).

## 데이터/스캔 계약
- DB 파일: `savelocation.db` (실행 경로 기준 상대).
- 스키마: `savelocation(name,type,filename,path)`.
- PK/UNIQUE 없음, 업데이트/삭제는 내부적으로 SQLite `rowid` 사용.
- `zip` 감지: Glob 매치 수.
- `reg` 감지: 규칙(`\\` 종료 여부)으로 키/값 타깃 판별 후 존재 여부 확인.

## 빌드/검증 절차
1. `go build ./...`
2. `fyne package --target windows --source-dir . --icon assets\app-icon.png --name game-save-backup-manager`
3. `New-Item -ItemType Directory -Force out | Out-Null`
4. `Move-Item -Force game-save-backup-manager.exe out\game-save-backup-manager.exe`

## 환경 특이사항
- 이 환경은 `go build` 시 캐시 경로 권한 이슈가 자주 발생한다.
- 실패 시 동일 명령을 권한 상승으로 재실행하면 통과되는 경우가 대부분.
- 현재 프로젝트는 Git 저장소(사용자 생성 완료).
