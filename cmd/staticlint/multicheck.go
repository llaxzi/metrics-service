// package main реализует пользовательский набор линтеров для статического анализа Go-кода,
// включая встроенные анализаторы из пакета go/analysis, сторонние анализаторы из staticcheck,
// а также пользовательский анализатор.
//
// Запускать через:
//
//	go vet -vettool=./staticlint ./...
//
// Анализаторы, добавленные в этот набор:
//
//   - printf.Analyzer:
//     Проверяет соответствие форматов в fmt.Printf-подобных функциях. Например, формат "%d"
//     требует аргумента типа int.
//
//   - shadow.Analyzer:
//     Находит случаи затенения переменных — когда переменная с тем же именем объявляется
//     повторно внутри вложенного блока, скрывая внешнюю переменную.
//
//   - structtag.Analyzer:
//     Проверяет корректность синтаксиса struct-тегов (например, `json:"name,omitempty"`).
//     Предупреждает о пробелах, дублировании и других ошибках.
//
//   - shift.Analyzer:
//     Ищет сдвиги бит, которые могут привести к потере значений (например, отрицательные или большие сдвиги).
//
//   - copylock.Analyzer:
//     Предупреждает, если mutex или другой sync.тип передаётся по значению, что может привести к гонке.
//
//   - unreachable.Analyzer:
//     Находит недостижимый код — код, который никогда не будет выполнен (например, после return).
//
//   - assign.Analyzer:
//     Ищет множественные присваивания переменной в пределах одного выражения (что может быть ошибкой).
//
//   - analyzer.Analyzer:
//     Пользовательский анализатор, реализованный в metrics-service/cmd/staticlint/analyzer.
//     Назначение зависит от твоей реализации — может проверять бизнес-ошибки, архитектурные ограничения и т.д.
//
// Дополнительно, программа подключает анализаторы из пакета staticcheck.io, включая:
//
//   - SAxxxx (все анализаторы, начинающиеся на "SA"):
//     Семантические проверки: подозрительные конструкции, возможные ошибки, дублирование, пустые блоки и т.д.
//     Примеры: SA1000 — неверное использование time.Parse, SA4006 — неиспользуемое значение.
//
//   - ST1000:
//     Проверка имени пакета: оно должно соответствовать имени директории (базовая рекомендация по стилю).
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/staticcheck"
	"metrics-service/cmd/staticlint/analyzer"
)

func main() {
	checks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		shift.Analyzer,
		copylock.Analyzer,
		// На мой выбор
		unreachable.Analyzer,
		assign.Analyzer,
		// Собственный анализатор
		analyzer.Analyzer,
	}

	// Добавляем анализаторы SA из staticcheck и один из анализаторов других классов staticcheck.
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer != nil && (v.Analyzer.Name[:2] == "SA" || v.Analyzer.Name == "ST1000") {
			checks = append(checks, v.Analyzer)
		}
	}

	multichecker.Main(checks...)

}
