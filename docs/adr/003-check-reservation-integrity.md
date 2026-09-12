# ADR-003: 二重予約の検知として、定期的にDBにクエリを投げて確認

## Context
二重予約に対する Correctness SLIを定義するために、DBにクエリを投げるコードが必要


## 参考資料

https://pkg.go.dev/go.opentelemetry.io/otel/metrics
https://uptrace.dev/opentelemetry/metrics

