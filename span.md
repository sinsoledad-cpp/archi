# OpenTelemetry Span 常用函数 (Go)

`Span` 代表分布式链路中的一个工作单元（如一次函数执行、一次 HTTP 请求或一次数据库查询）。以下是 `go.opentelemetry.io/otel/trace.Span` 接口最常用的函数，按功能分类。

## 1. Span 生命周期

这些函数用于创建和结束一个 Span。

### `tracer.Start(ctx, "span-name")`

这不是 `Span` 接口的函数，而是 `Tracer` 的函数，但它是**创建 Span 的起点**。

* **作用**：开始一个新的 Span，并将其作为当前 Span 注入到返回的 `context.Context` 中。
* **用法**：
    ```go
    // tracer 是一个 trace.Tracer 实例
    ctx, span := tracer.Start(ctx, "my-operation")
    ```

### `span.End()`

* **作用**：标记 Span 结束。这是**必须调用**的函数，否则 Span 不会被导出。它会计算 Span 的总持续时间。
* **用法**：
    ```go
    defer span.End() // 99% 的情况下都应该使用 defer 来确保它被调用
    ```

## 2. 记录错误与状态

这是标记 Span 是否成功的最重要方法。

### `span.RecordError(err error)`

* **作用**：记录一个 Go 的 `error` 对象。这会自动将错误信息（包括堆栈跟踪）添加为 Span 上的一个特殊 "exception" 事件。
* **用法**：
    ```go
    if err != nil {
        span.RecordError(err)
        // ...
    }
    ```

### `span.SetStatus(code codes.Code, description string)`

* **作用**：设置 Span 的最终状态。
* **重要**：`RecordError` **不会**自动将 Span 状态设置为 "Error"。您必须**手动调用 `SetStatus`**，这样 Jaeger、Zipkin 等 UI 才会将此 Span 标红。
* **用法**：
    ```go
    import "go.opentelemetry.io/otel/codes"

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error()) // 强烈建议与 RecordError 一起使用
    } else {
        // 默认状态是 Unset (等同于 OK)，但也可以显式设置
        span.SetStatus(codes.Ok, "operation successful")
    }
    ```

## 3. 添加元数据（属性）

“属性”（Attributes）是可供查询、筛选和聚合的键值对“标签”。

### `span.SetAttributes(attributes ...attribute.KeyValue)`

* **作用**：为 Span 添加一个或多个属性。如果 Key 已存在，则会**覆盖**。这是添加业务上下文（如 `user.id`, `order.id`）最常用的方法。
* **用法**：
    ```go
    import "go.opentelemetry.io/otel/attribute"

    span.SetAttributes(
        attribute.String("tpl.id", "template-123"),
        attribute.Int("sms.count", 5),
        attribute.Bool("is.priority", true),
    )
    ```

## 4. 添加事件（日志）

“事件”（Events）是带时间戳的“结构化日志”，附加到 Span 的时间轴上。

### `span.AddEvent(name string, options ...trace.EventOption)`

* **作用**：在 Span 的生命周期内记录一个特定的时间点事件。
* **与 `SetAttributes` 的区别**：
    * `SetAttributes`：是 Span 的**静态属性**（标签）。
    * `AddEvent`：是 Span 期间发生的**动态事件**（日志）。
* **用法**：
    ```go
    span.AddEvent("开始请求短信网关")
    // ... 执行一些操作 ...
    span.AddEvent("网关响应完毕")
    
    // 也可以为事件添加属性
    span.AddEvent("收到响应", trace.WithAttributes(
        attribute.Int("http.status_code", 200),
    ))
    ```

## 5. 性能优化

### `span.IsRecording()`

* **作用**：返回一个布尔值，指示此 Span **是否真的在被记录**。如果 OpenTelemetry 的采样器（Sampler）决定丢弃此链路，它将返回 `false`。
* **用法**：用于避免在 Span 被丢弃时执行昂贵的计算来获取属性。
    ```go
    if span.IsRecording() {
        // 假设 computeExpensiveData() 是一个耗时操作
        data := computeExpensiveData() 
        span.SetAttributes(attribute.String("expensive.data", data))
    }
    ```