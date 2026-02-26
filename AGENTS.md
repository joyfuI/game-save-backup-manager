# AGENTS.md

## 목적
새 세션 에이전트가 코드 전체를 다시 파악하지 않고도, 작업 중 놓치기 쉬운 규칙과 합의사항을 빠르게 이어받도록 한다.

## 하드 규칙 (중요)
- exe 산출물은 프로젝트 루트의 `game-save-backup-manager.exe`.
- Windows 패키징은 `FyneApp.toml` 메타데이터 기반 `fyne package -os windows`를 사용.
- 배포 패키징은 `fyne package -os windows -release`를 기본으로 사용.
- `reg` 경로 해석 규칙은 고정:
  - 경로가 `\\`로 끝나면 키 전체 백업/검증 대상.
  - 경로가 `\\`로 끝나지 않으면 마지막 세그먼트를 값 이름으로 간주해 값 단일 백업/검증 대상.
- 위 `reg` 규칙은 스캔/백업/입력검증에 동일하게 적용해야 한다.

## 현재 UI 계약 (2026-02-26 기준)
- 메인 상단: 왼쪽 `스캔`, 오른쪽 `설정`.
- `설정` 버튼은 별도 창이 아니라 다이얼로그를 연다.
- 설정 다이얼로그의 현재 구성:
  - `Ubisoft Connect 설치 경로` 입력
  - `Ubisoft Connect USER ID` 입력 (기본값 빈 문자열)
  - `폴더 선택` 버튼(Fyne 폴더 다이얼로그)
  - `DB 관리` 버튼
  - `DB 관리` 아래 버튼: `저장` / `취소` (가로 1:1 너비)
- 설정 다이얼로그는 너비에 따라 `Ubisoft Connect 설치 경로` 라벨/입력 UI가 1줄 또는 2줄로 반응형 배치된다.
- `DB 관리` 버튼을 누르면 기존 DB 관리 창(등록/수정/삭제/새로고침 + 테이블)이 열린다.
- 메인 목록은 전체 DB가 아니라 감지된 세이브만 리스트로 표시한다.
- 메인 목록 행 구성: 게임명 + `백업` + `경로 열기`.
- 메인 창 종료 시, 열려 있는 `DB 관리` 창도 함께 닫힌다.

## 놓치기 쉬운 동작
- `reg` 타입의 `경로 열기`는 파일 탐색기가 아니라 레지스트리 편집기를 연다.
- 레지스트리 편집기 실행이 일반 권한으로 실패하면 `RunAs`로 UAC 승격 재시도한다.
- `zip` 백업은 ZIP 내부에서 디렉터리 계층을 보존하며, 환경변수 기반 논리 경로(예: `%APPDATA%`)도 유지한다.

## 데이터/스캔 계약
- DB 파일: `savelocation.db` (실행 경로 기준 상대).
- 스키마: `savelocation(name,type,filename,path)`.
- PK/UNIQUE 없음. 업데이트/삭제는 SQLite `rowid` 사용.
- `zip` 감지: Glob 매치 수 기준.
- `reg` 감지: 위 경로 규칙으로 키/값 타깃을 판별한 뒤 존재 여부 확인.
- 세이브 경로 플레이스홀더 규칙:
  - `{{ubisoftconnect-folder}}` => 설정의 `Ubisoft Connect 설치 경로`
  - `{{ubisoftconnect-user-id}}` => 설정의 `Ubisoft Connect USER ID`
  - 스캔/실제 파일 접근 시에는 플레이스홀더를 설정값으로 치환해 사용
  - ZIP 내부 논리 경로는 플레이스홀더 원문을 유지
  - 치환에 필요한 설정값이 비어 있으면 해당 항목은 미감지로 처리될 수 있음

## 설정 파일 계약
- 파일 형식: INI
- 파일명: 실행 파일명 기반 (`<exe basename>.ini`)
- 저장 위치: 실행 파일과 같은 디렉터리
- 키: `ubisoft_connect_path`
- 키: `ubisoft_connect_user_id`
- 기본값(ini 없을 때): `%PROGRAMFILES(X86)%\Ubisoft\Ubisoft Game Launcher`
  - UI 표시 시 환경변수를 실제 경로로 치환한 값 사용
  - 저장 시에도 치환된 실제 경로를 저장
- `ubisoft_connect_user_id` 기본값: 빈 문자열
- 프로그램 시작 시 ini가 없으면 기본값으로 자동 생성한다.

## 빌드/검증 절차
1. `go build ./...`
2. `fyne package -os windows`
3. 배포 빌드 시 `fyne package -os windows -release`

## 환경 특이사항
- 이 환경은 `go build` 시 캐시 경로 권한 이슈가 간헐적으로 발생한다.
- 실패 시 동일 명령을 권한 상승으로 재실행하면 통과되는 경우가 많다.
- 현재 프로젝트는 Git 저장소다.
