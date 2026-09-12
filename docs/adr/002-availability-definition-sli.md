# ADR-002: このシステムにおけるSLIの定義を記載

## Context
本番運用を想定した場合、システムに対しSLIの定義が必要
ユーザからのAPIリクエストがあることを想定

## Decision

SLIとして、可用性、レイテンシ、エラー率でまずは考える。
ユーザ視点として、システムに求めることは以下

* 座席の表示や試合の表示などのリクエストに対するレスポンスが早いこと
* ユーザからのリクエストに対し、どれくらいのレスポンスが成功するか
* 予約を購入するときにきちんと成功すること
* 二重予約、二重決済などが起きないこと
* サービス自体のダウン時間が極力少ないこと
* 多くのユーザがアクセスしたときにダウン/通信遅延が起きないこと

CUJ (Critical User Journey)の定義
1. 試合を探す (GET /api/agmes, GET /api/games/{id})
2. 座席を選ぶ (GET /api/games/{id}/seats)
3. 予約する (POST /api/reservations)
4. 決済する (POST /api/reservations/{id}/purchase)
5. 予約履歴を見る (GET /api/reservations, GET /api/reservations/{id})

SLI定義
 CUJ ごとにSLIを定義する。
 可用性はそれぞれのリクエストに対し、
 「成功したリクエスト数」 / 全リクエスト数

* 測定ウィンドウは24時間単位
* レイテンシはサーバ側の処理時間

※成功したリクエストとは、4xx,5xx以外のリクエスト全般
※分母のリクエストにおいて、ユーザ起因(4xx)のリクエストは除外

## Consequences
 - Prometheus等の可視化基盤が必要
 - 429のRate Limitは対応していないので、可用性の値は少しズレがある
 - 

## 参考資料

https://sre.google/workbook/implementing-slos/
https://www.linkedin.com/pulse/implementing-sli-slo-like-sre-practical-guide-beginners-harsur-btraf/


