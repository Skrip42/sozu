# sozu

Итак sozu. Это отпугивалка оленей, которая поможет вам собрать ваши данные в кучку.

## Базовый функционал

На текущий момент реализовано 2 режима работы: буфер и агрегатор

### Буфер
Буфер накапливает масив данных и отдает его по запросу.

Чтобы создать фабрику буферов, выполните `fabric := sozu.NewBuffer[type]()`
И затем создайте любое количество буферов вызвав
`output, flush := fabric.Create(ctx, input)` где

- `input` канал типа type, куда предпологается передавать входные данные.
- `ctx` контекст для экстреного завершения работы буфера
- `output` канал типа []type с масивом выходных данных
- `flush` функция принуждающая буфер отдать данные

Пример кода с использованием буфера можно посмотреть [тут](example/buffer/buffer.go)

:warning: Для коректного завершения работы буфера нужно азкрыть входной канал,
это позволит буферу коректно отдать данные прежде чем завершить работу
Закрытие контекста привидет к экстреной остановке работы буфера с потерей всех данных находящихся в нем

### Агрегатор
Агрегатор работает похожим образом, но позволяет сэкономить на алокации масива,
в случае если вам необходимо просто выполнить простую операцию с данными.
Например если вам нужно хранить только максимальный элемент

Чтобы создать фабрику агрегаторов выполните `fabric := sozu.NewAggregator[type](aggregateFunc)` где
`aggregateFunc` это функция с сигнатурой `func(type, type) type`
Затем вы можете создать любое количество агрегаторов вызвав
`output, flush := fabric.Create(ctx, input)` где

- `input` канал типа type, куда предпологается передавать входные данные.
- `ctx` контекст для экстреного завершения работы буфера
- `output` канал типа []type с масивом выходных данных
- `flush` функция принуждающая буфер отдать данные

