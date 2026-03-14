# Remote Journal Frontend Regression

Date: 2026-03-14

Purpose: close `JRV-040` by recording the minimal live frontend regression
against the remote single-db profile started with `./start.sh dev remote`.

## Test Objects

- Primary user: `rv-front-1773475922` (`biz_user.id=6`)
- Paper sample: `RV temporary paper 20260314` (`biz_paper.id=1`)
- Primary rating: `biz_rating.id=1` by user `6`
- Paper flag: `biz_flag.id=1` on `paper.id=1`
- Reporter user: `rv-reporter-1773476041` (`biz_user.id=7`)
- Rater user: `rv-rater-1773476041` (`biz_user.id=8`)
- Third-party rating: `biz_rating.id=2` by user `8`
- Third-party rating flag: `biz_flag.id=3` on `rating.id=2`

## Request And Response Summary

- `POST /api/v1/user/register`
  Request: `{"username":"rv-front-1773475922","email":"codex+rv-front-1773475922@example.invalid"}`
  Response: `{"id":6}`
- `POST /api/v1/user/login`
  Request: `{"username":"rv-front-1773475922"}`
  Response: `user_id=6`, JWT issued, `admin_permissions=[]`
- `GET /api/v1/user/info`
  Response: `id=6`, `username=rv-front-1773475922`, `role=0`
- `GET /api/v1/papers?page=1&page_size=5`
  Response: `total=1`, first paper `id=1`
- `GET /api/v1/papers/1`
  Response: `id=1`, `keywords=rv_tmp_keyword_20260314,remote-validation-20260314`
- `GET /api/v1/papers/search?query=RV%20temporary&page=1&page_size=5`
  Response: `total=1`, `first_id=1`, `engine=fulltext`
- `POST /api/v1/papers/1/rate`
  Request: `{"score":7,"comment":"rv-rate-1773475922"}`
  Response: `{"success":true,"message":"rating submitted"}`
- `GET /api/v1/papers/1/ratings`
  Response: `total=1`, `first_rating_id=1`, `avg_score=7`
- `POST /api/v1/papers/1/flag`
  Request: `{"reason":"spam","detail":"rv-paper-flag-1773475922"}`
  Response: `flag_id=1`, `target_type=paper`, `pending_count=1`
- `POST /api/v1/ratings/1/flag`
  Request: self-flag attempt by user `6`
  Response: HTTP `200` with `success=false`, message `不能举报自己 / Cannot flag yourself`
- `POST /api/v1/user/register` again with the same username/email
  Response: HTTP `500`, `username already taken`
- `POST /api/v1/papers/1/rate` by user `8`
  Request: `{"score":5,"comment":"rv-second-rate-1773476041"}`
  Response: `{"success":true,"message":"rating submitted"}`
- `POST /api/v1/ratings/2/flag` by user `7`
  Request: `{"reason":"spam","detail":"rv-third-party-flag-1773476041"}`
  Response: `flag_id=3`, `target_type=rating`, `pending_count=1`

## Remote DB Evidence

- `biz_user.id IN (6,7,8)` exists for the disposable frontend users.
- `biz_rating.id=1` stores `paper_id=1`, `user_id=6`, `score=7`,
  `comment=rv-rate-1773475922`.
- `biz_rating.id=2` stores `paper_id=1`, `user_id=8`, `score=5`,
  `comment=rv-second-rate-1773476041`.
- `biz_flag.id=1` stores a pending paper flag on `paper.id=1` from reporter `6`.
- `biz_flag.id=3` stores a pending rating flag on `rating.id=2` from reporter `7`.

## Review Notes

- The successful chain reuses the same live remote-profile services from
  `JRV-030`, so the requests exercise `user.rpc`, `paper.rpc`, `rating.rpc`,
  `news.rpc`, and `journal-api` against the remote journal DSN.
- The abnormal path coverage for this issue is:
  self-flag rejection on `rating.id=1` and duplicate registration rejection for
  `rv-front-1773475922`.
