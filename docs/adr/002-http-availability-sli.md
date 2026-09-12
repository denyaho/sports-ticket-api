# ADR-001: HTTP AvailabilityをSLIとして採用

## Context
本番運用を想定して、クライアントからの通信に対しSLIを定義

## Decision
Google SREの考え方を参考に
Availabilityはリクエストの成功率という観点で定義する
Availability = (有効なリクエスト数 - サービス起因のエラー数) / 有効なリクエスト数　となる。

有効なリクエスト数


## 参考資料

https://www.linkedin.com/pulse/implementing-sli-slo-like-sre-practical-guide-beginners-harsur-btraf/
