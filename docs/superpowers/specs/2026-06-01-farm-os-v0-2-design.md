# FARM.OS v0.2 — Design Spec

**Data:** 2026-06-01  
**Base:** `farm-idle v0.1` já funcional em Go + Bubble Tea + Lipgloss  
**Objetivo:** reduzir microgerenciamento e transformar loop em gestão operacional

---

## Visão

`v0.2` adiciona automação e feedback operacional sem explodir complexidade.

Meta do release:

- menos cliques
- zero softlock por falta de sementes
- feedback visual melhor do campo
- base correta para trabalhadores e gargalos

Loop desejado:

```text
dinheiro → sementes → plantio automático → crescimento → colheita → transporte/venda → expansão → mais capacidade
```

Diferença central para `v0.1`:

- `v0.1` ainda depende de ajustes manuais frequentes
- `v0.2` sustenta produção sozinho e começa a expor gargalos de throughput

---

## Escopo

### Dentro de `v0.2`

#### Bloco A — Fundação operacional

1. compra manual de sementes em lote
2. auto-venda corrigida para disparar em `>=`
3. header com `Receita/min`
4. mini-grid real do campo
5. auto-compra com mínimo programável
6. nova ação de menu para configurar auto-compra

#### Bloco B — Automação administrável

1. contratação de operadores
2. distribuição manual entre `Plantio`, `Colheita`, `Transporte`
3. throughput limitado por capacidade de operadores
4. painel de automação
5. cálculo de gargalo atual
6. cálculo de eficiência por etapa e global

### Fora de `v0.2`

- clima
- NPC
- mercado dinâmico
- múltiplas culturas
- salários
- IA de redistribuição
- múltiplos tipos de trabalhador
- fila de jobs detalhada

---

## Objetivos de design

1. manter engine simples e testável
2. evitar novos estados invisíveis para o jogador
3. introduzir automação por camadas
4. fazer painel refletir sistema real, não número fake
5. permitir iteração rápida para `v0.3`

---

## Arquitetura proposta

### Estratégia

Seguir 2 fases dentro do mesmo release:

- `Bloco A`: corrige economia e UX básica sem mudar arquitetura profundamente
- `Bloco B`: adiciona camada operacional de capacidade por etapa

### Abordagem escolhida para trabalhadores

`Operadores por etapa com capacidade simples`.

Cada tick terá capacidade limitada em:

- `Plantio`
- `Colheita`
- `Transporte`

Operadores aumentam capacidade. Gargalo sai da etapa com menor capacidade efetiva diante da demanda atual.

Essa abordagem foi escolhida porque:

- encaixa no painel desejado
- é simples de explicar
- é testável sem fila complexa de jobs
- evita overengineering de `v0.2`

---

## Modelo de domínio

### Estado novo ou alterado

O `Model` atual será expandido.

#### Bloco A

- `SeedsPerPurchase int` — fixo em `5` neste release, mas armazenado no modelo para evitar número mágico espalhado
- `AutoBuyEnabled bool`
- `AutoBuyMinimum int`
- `RevenueWindow []float64` ou estrutura equivalente para cálculo de `Receita/min`
- `RecentRevenue float64` como valor derivado pronto para renderer

#### Bloco B

- `OperatorsTotal int`
- `OperatorsPlanting int`
- `OperatorsHarvest int`
- `OperatorsTransport int`
- `AutomationMetrics` ou campos equivalentes para:
  - `PlantingUtilization float64`
  - `HarvestUtilization float64`
  - `TransportUtilization float64`
  - `Bottleneck string`
  - `GlobalEfficiency float64`

### Invariantes

- `OperatorsPlanting + OperatorsHarvest + OperatorsTransport <= OperatorsTotal`
- `AutoBuyMinimum >= 0`
- `Seeds >= 0`
- `Stock >= 0`
- capacidades nunca negativas
- eficiências entre `0` e `1`

### Persistência

Todos os novos campos operacionais devem entrar em `SaveData`, exceto métricas puramente derivadas do tick atual que podem ser recalculadas.

Persistir:

- config de auto-compra
- total de operadores
- distribuição de operadores
- dados suficientes para restaurar receita recente com fallback seguro

Fallback aceitável:

- ao carregar save antigo sem os novos campos, usar defaults compatíveis e não quebrar save legado
- default de `AutoBuyMinimum` para save antigo: `5`

---

## Regras de negócio

## Bloco A

### A1. Compra manual em lote

Regra:

- ação de compra passa a comprar `5 sementes` por `$10`

Resultado esperado:

- reduz risco de softlock
- reduz spam de tecla

Observação:

- manter nome simples no menu, mas o label deve deixar claro que é lote

### A2. Auto-venda

Troca:

```go
stock > threshold
```

por:

```go
stock >= threshold
```

Efeito:

- ao atingir limite, venda dispara imediatamente

### A3. Receita/min

Substitui `Lucro/min`.

Definição:

- medir receita recente de venda em janela móvel de 60 ticks
- contar apenas dinheiro gerado por venda automática ou venda manual
- não contar gasto com sementes, expansão ou upgrades como receita negativa

Justificativa:

- é métrica operacional, não patrimonial
- evita distorção visual por compras e upgrades

### A4. Mini-grid do campo

Renderer mostrará grade baseada em `Plants[]`.

Mapeamento visual:

- `empty` → `□`
- `planted` → `·`
- `growing` → `▓`
- `ready` → `█`

Layout:

- grid compacto com quebra de linha previsível
- preferir largura fixa pequena, ex.: `4`, `5` ou derivada do `FieldSize`
- abaixo do grid, manter resumo textual `ativos/total`

### A5. Auto-compra

Configuração:

- `Auto-compra` ligada/desligada
- mínimo programável

Regra:

- se `AutoBuyEnabled == true`
- e `Seeds < AutoBuyMinimum`
- então comprar automaticamente `5 sementes` por `$10`
- só comprar se houver dinheiro suficiente

Objetivo:

- sustentar plantio automático
- evitar loop morto por falta de semente

### A6. Menu

Nova ação:

- `[6] Config auto-compra`

UX mínima:

- comportamento similar ao input inline da auto-venda
- entrar em modo config e digitar mínimo na mesma linha

Decisão recomendada:

- usar mesmo padrão do input inline já existente
- exibir algo como `"[6] Auto-compra min: 5_"` durante edição

---

## Bloco B

### B1. Operadores

Primeira função dos operadores:

- aumentar throughput

Não fazem:

- decisões inteligentes
- auto-balanceamento
- salários

### B2. Distribuição manual

Player aloca manualmente entre:

- `Plantio`
- `Colheita`
- `Transporte`

O jogo nunca redistribui sozinho em `v0.2`.

### B3. Capacidade por etapa

Cada etapa terá fórmula de capacidade por tick.

Forma da regra:

- `capacidade base + bonus por operador`

O valor exato pode ser ajustado no plano, mas a arquitetura assume:

- plantio pode ser limitado por operadores
- colheita pode acumular plantas prontas se faltar capacidade
- transporte pode acumular estoque acima do que seria vendido por tick

### B4. Gargalo

`Gargalo` é etapa com menor razão entre capacidade disponível e demanda atual.

Exemplos:

- muitas plantas prontas e poucos operadores de colheita → gargalo `Colheita`
- estoque acumulado e pouco transporte → gargalo `Transporte`
- muitos slots vazios, semente disponível e pouco plantio → gargalo `Plantio`

### B5. Eficiência

Cada etapa mostrará percentual de utilização.

Definição sugerida:

```text
utilização = demanda_atendida / capacidade_disponível
```

ou, equivalentemente, quando capacidade > 0:

```text
utilização = demanda / capacidade, limitada a 100%
```

Critério do release:

- percentual precisa ser consistente
- não precisa ser economicamente perfeito
- precisa apontar gargalo real

### B6. Painel de automação

Novo painel:

```text
┌──────── AUTOMAÇÃO ────────┐
│ Operadores: 4             │
│                           │
│ Plantio     ██████ 80%    │
│ Colheita    ████░░ 52%    │
│ Transporte  ██░░░░ 30%    │
├───────────────────────────┤
│ Gargalo: Transporte       │
└───────────────────────────┘
```

Esse painel deve refletir dados do engine, não heurística do renderer.

---

## Fluxo do tick

### Tick atual

Hoje `Tick()` faz quase tudo em passagem única.

### Tick proposto para `v0.2`

Separar em etapas lógicas:

1. auto-compra de sementes
2. plantio até limite permitido
3. avanço de crescimento
4. colheita até limite permitido
5. transporte/venda até limite permitido
6. atualização de métricas de receita/min
7. atualização de métricas de automação
8. avanço de tempo (`TickCount`, `Day`)

### Compatibilidade de fases

No `Bloco A`:

- limites de throughput ainda podem ser “infinito prático”
- mas estrutura do código já deve começar a ficar separada por etapas

No `Bloco B`:

- mesmas etapas passam a respeitar capacidade

Isso evita reescrever engine duas vezes.

---

## Renderer e UX

### Header

Trocar:

- `Lucro/min`

por:

- `Receita/min`

Header ideal:

```text
Dia X  │  Receita/min: $42  │  Auto-venda: ≤5  │  Auto-compra: min 5
```

### Campo

Adicionar mini-grid sem remover totalmente resumo textual.

Objetivo:

- sensação de fazenda
- leitura rápida de estado real

### Menu de ações

Bloco A deve adicionar ao menos:

- compra de sementes em lote
- config de auto-compra

Bloco B deve adicionar depois:

- contratar operador
- mover operador para plantio
- mover operador para colheita
- mover operador para transporte

Observação:

- se o menu ficar longo demais, o plano pode dividir ações em submenu ou paginação simples
- isso é decisão de implementação, não requisito da spec

### Logs operacionais

Expandir logs para refletir automação de forma mais agregada quando possível.

Exemplos úteis:

- `+5 sementes`
- `auto-compra: +5 sementes`
- `3 colheitas`
- `venda $20`
- `operador contratado`
- `alocação: colheita +1`

Evitar spam por evento microscópico quando houver throughput alto.

---

## Migração e compatibilidade

### Saves antigos

`Load()` deve aceitar saves de `v0.1`.

Defaults esperados:

- `AutoBuyEnabled = false`
- `AutoBuyMinimum = 5`
- `OperatorsTotal = 0`
- distribuições = `0`
- métricas derivadas recalculadas

### Falhas de configuração

Se estado salvo estiver inconsistente:

- clamp de valores negativos
- clamp de alocação acima do total
- fallback para config segura

---

## Error handling

- compra automática sem dinheiro: não quebra tick, apenas não compra
- alocação inválida de operadores: clamp silencioso para estado válido
- save antigo incompleto: carregar com default
- renderer sem métricas disponíveis: mostrar zeros/defaults, nunca quebrar

---

## Estratégia de testes

### Bloco A

Cobrir:

1. compra manual rende `+5 sementes`
2. auto-venda dispara em `>=`
3. receita/min considera apenas receita recente
4. auto-compra dispara quando `Seeds < minimum`
5. auto-compra não dispara sem dinheiro
6. save/load preserva config de auto-compra
7. mini-grid renderiza slots corretamente

### Bloco B

Cobrir:

1. contratação incrementa total
2. alocação manual respeita total disponível
3. capacidade limita plantio
4. capacidade limita colheita
5. capacidade limita transporte
6. gargalo muda conforme distribuição
7. renderer mostra percentuais e gargalo corretos
8. save/load preserva operadores e alocação

### Smoke manual

Critérios de aceite:

- após 30 minutos, produção segue crescendo
- jogo não entra em softlock
- jogador toma decisões de config/alocação
- painel de automação mostra gargalo legível

---

## Plano de execução recomendado

### Fase 1 — Bloco A

1. lote manual de sementes
2. fix auto-venda `>=`
3. receita/min
4. auto-compra + persistência
5. nova ação `[6]`
6. mini-grid
7. smoke e balanceamento inicial

### Fase 2 — Bloco B

1. modelo de operadores
2. contratação
3. distribuição manual
4. refatorar tick por etapas
5. throughput por etapa
6. métricas de eficiência
7. gargalo
8. painel de automação
9. smoke final

---

## Critério de sucesso

`v0.2` está bom quando:

- produção não trava por semente
- menu exige menos microgerenciamento
- campo comunica estado visualmente
- receita/min reflete operação real
- operadores mudam throughput perceptivelmente
- gargalo aparece de forma compreensível
- jogador sente que está administrando sistema, não apertando botão
