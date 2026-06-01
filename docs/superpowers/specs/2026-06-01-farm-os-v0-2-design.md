# FARM.OS v0.2 — Design Spec

**Data:** 2026-06-01  
**Base:** `farm-idle v0.1` já funcional em Go + Bubble Tea + Lipgloss  
**Objetivo:** reduzir microgerenciamento sem mudar o modelo inteiro do jogo

---

## Visão

`v0.2` é um release de fundação operacional.

Ele não introduz operadores, gargalos nem painel de automação. Esses itens ficam para `v0.3`.

O foco agora é:

- remover risco de softlock
- reduzir clique repetitivo
- melhorar sinal operacional
- melhorar leitura visual do campo

Loop desejado em `v0.2`:

```text
dinheiro → sementes → plantio automático → crescimento → colheita automática → venda → expansão
```

Diferença central para `v0.1`:

- `v0.1` funciona, mas ainda exige microgerenciamento e tem risco econômico ruim
- `v0.2` sustenta produção com menos atrito e prepara terreno para automação pesada depois

---

## Escopo

### Dentro de `v0.2`

1. compra manual de sementes em lote
2. auto-venda corrigida para disparar em `>=`
3. header com `Receita/min`
4. mini-grid real do campo
5. auto-compra com mínimo programável
6. nova ação de menu para configurar auto-compra
7. limite de gasto da auto-compra

### Fora de `v0.2`

- operadores
- throughput por etapa
- gargalos
- painel de automação
- clima
- NPC
- mercado dinâmico
- múltiplas culturas
- salários
- IA de redistribuição

### Roadmap resultante

#### `v0.2`

- lote sementes
- auto-compra
- auto-venda
- receita/min
- mini-grid

#### `v0.3`

- operadores
- throughput
- gargalos
- painel de automação

---

## Objetivos de design

1. manter engine simples e testável
2. evitar overengineering cedo
3. não automatizar economia quebrada
4. fazer renderer comunicar melhor o estado real
5. preparar engine para futura separação por etapas sem exigir isso agora

---

## Arquitetura proposta

### Estratégia

`v0.2` continua usando o modelo atual de engine com evolução incremental.

Não haverá nova camada de jobs, nem sistema de operadores, nem throughput limitado por etapa neste release.

Mudanças devem preservar regra:

- `engine` = sem IO
- `renderer` = sem regra
- `persistência` = simples

### Princípio de implementação

Mesmo sem reescrever o tick inteiro em `v0.2`, o código novo deve evitar acoplamento desnecessário e não bloquear a futura migração para sistema de throughput em `v0.3`.

---

## Modelo de domínio

### Estado novo ou alterado

O `Model` atual será expandido apenas para o necessário em `v0.2`.

### Campos novos

- `SeedsPerPurchase int`
- `AutoBuyEnabled bool`
- `AutoBuyMinimum int`
- `AutoBuyMaxCashFraction float64`
- `RevenueTracker RevenueTracker`
- `RecentRevenue float64`

### RevenueTracker

Não usar slice dinâmica.

Estrutura desejada:

```go
type RevenueTracker struct {
    Buckets [60]float64
    Cursor  int
    Total   float64
}
```

Regras:

- um bucket por tick recente
- avanço circular
- atualização em `O(1)`
- `RecentRevenue` pode ser derivado do `Total`

### Invariantes

- `SeedsPerPurchase == 5` em `v0.2`
- `AutoBuyMinimum >= 0`
- `0 <= AutoBuyMaxCashFraction <= 1`
- `Seeds >= 0`
- `Stock >= 0`
- `RevenueTracker.Total >= 0`

### Persistência

Persistir:

- config de auto-compra
- `SeedsPerPurchase`
- `RevenueTracker`

Se `RevenueTracker` completo for considerado pesado demais para save, é aceitável persistir apenas o bastante para recomeçar zerado após load, desde que:

- isso seja explícito no código
- não quebre o jogo
- o header mostre valor consistente logo após alguns ticks

### Compatibilidade com saves antigos

Ao carregar save de `v0.1`, usar:

- `SeedsPerPurchase = 5`
- `AutoBuyEnabled = false`
- `AutoBuyMinimum = 5`
- `AutoBuyMaxCashFraction = 0.30`
- `RevenueTracker` zerado
- `RecentRevenue = 0`

---

## Regras de negócio

### A1. Compra manual em lote

Regra:

- ação de compra passa a comprar `5 sementes` por `$10`

Observação:

- essa é escolha de design do release, mesmo que no futuro o balanceamento mude

Resultado:

- menor risco de softlock
- menor spam de input

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

- medir apenas receita de vendas recentes
- janela móvel de `60` ticks
- venda manual e automática contam
- compra de semente, expansão e upgrade não reduzem essa métrica

Justificativa:

- é uma métrica operacional, não patrimonial
- evita sinal enganoso

### A4. Mini-grid do campo

Renderer mostrará grade baseada em `Plants[]`.

Mapeamento visual:

- `empty` → `□`
- `planted` → `·`
- `growing` → `▓`
- `ready` → `█`

Regra:

- grid representa slots reais
- resumo textual continua existindo abaixo ou ao lado

Nota futura:

- manter `Plants []PlantSlot` como estrutura principal
- `ActiveSlots` fica apenas como possível otimização futura, não entra neste release

### A5. Auto-compra

Configuração:

- `Auto-compra` ligada/desligada
- mínimo programável

Regra de disparo:

- se `AutoBuyEnabled == true`
- e `Seeds < AutoBuyMinimum`
- e `Money >= 10`
- e `10 <= Money * AutoBuyMaxCashFraction`
- então comprar automaticamente `5 sementes` por `$10`

Valor de release:

- `AutoBuyMaxCashFraction = 0.30`

Objetivo:

- sustentar plantio automático
- evitar que reposição automática sabote upgrades e expansão cedo demais

### A6. Menu

Nova ação:

- `[6] Config auto-compra`

UX mínima:

- igual ao padrão inline da auto-venda
- digitação na mesma linha
- `Enter` confirma
- `Esc` cancela

Forma visual sugerida:

```text
[6] Auto-compra min: 5
```

Durante edição:

```text
[6] Auto-compra min: 5_
```

---

## Fluxo do tick

`v0.2` ainda pode usar tick único, mas com responsabilidades mais claras.

Ordem lógica:

1. tentar auto-compra
2. processar plantio automático já existente
3. avançar crescimento
4. colher plantas prontas
5. vender se `stock >= threshold`
6. atualizar `RevenueTracker`
7. atualizar `RecentRevenue`
8. avançar `TickCount` e `Day`

Importante:

- `v0.2` não precisa introduzir limite de throughput
- `v0.2` não precisa separar transporte de venda

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

Adicionar mini-grid sem remover resumo textual.

Objetivo:

- sensação de fazenda
- leitura instantânea do estado do campo

### Menu de ações

Em `v0.2`, menu precisa contemplar:

- compra de sementes em lote
- config de auto-venda
- config de auto-compra

Não adicionar ainda:

- contratar operador
- realocação operacional

### Logs

Logs devem refletir operações relevantes sem spam excessivo.

Exemplos úteis:

- `+5 sementes`
- `auto-compra: +5 sementes`
- `venda $20`

---

## Error handling

- auto-compra sem dinheiro suficiente: não compra, não quebra tick
- auto-compra bloqueada pelo limite de caixa: não compra, não quebra tick
- save antigo sem campos novos: carregar com defaults
- renderer sem receita recente suficiente: mostrar `0` de forma estável

---

## Estratégia de testes

Cobrir:

1. compra manual rende `+5 sementes`
2. compra manual custa `$10`
3. auto-venda dispara em `>=`
4. receita/min conta apenas vendas recentes
5. `RevenueTracker` gira em `O(1)` sem crescimento
6. auto-compra dispara quando `Seeds < minimum`
7. auto-compra não dispara com dinheiro insuficiente
8. auto-compra não dispara se custo exceder `30%` do caixa
9. save/load preserva config de auto-compra
10. save/load preserva tracker ou faz fallback explícito
11. mini-grid renderiza slots corretamente

### Smoke manual

Critérios de aceite:

- produção não trava por falta de semente
- auto-compra não suga caixa de forma burra
- receita/min reage a vendas reais
- campo fica visualmente mais legível

---

## Critério de sucesso

`v0.2` está bom quando:

- jogador para de lutar contra falta de semente
- loop exige menos clique
- header mostra sinal operacional útil
- campo comunica estado visualmente
- economia continua simples e estável
- base fica pronta para operadores em `v0.3`
