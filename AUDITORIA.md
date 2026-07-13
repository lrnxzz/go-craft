# go-craft — Plano de Ação (auditoria 2026-07-13)

Auditoria de 5 frentes (core/wire, nbt, mojang, runtime/codec, estrutura/testes) +
pesquisa sobre goroutine patterns, generics, protocolo MC e error handling.
Objetivo: mapear o que está fraco antes de partir pro world/block/entity model.
**Nada de código foi editado.** Prioridade por impacto real, não por esforço.

Calibração honesta: o **core de wire/protocolo é a parte forte** (DAG sem ciclos,
`go vet` limpo nos 5 módulos, `-race` limpo em core+codec, cobertura 83.6% core /
82.1% nbt, encoding append-based, bounds de frame/inflate sensatos). As fraquezas
se concentram em (1) alguns bugs de correção pontuais mas sérios, (2) robustez
contra input hostil, e (3) o *envelope* do projeto (publicável, testado nas bordas,
com storefront). Nenhum problema estrutural profundo — é polimento de alto nível.

---

## P0 — Bugs de correção (quebram comportamento real ou derrubam o processo)

**P0.1 — `FeatureFlags.ID()` errado quebra join real.**
`codec/v765/configuration.go:129` retorna `0x08`, mas Feature Flags é **`0x07`** no
protocolo 765; `0x08` é **Update Tags**. Confirmado direto no arquivo e na
minecraft-data. No join real o servidor manda os dois durante Configuration: o
Update Tags (0x08) cai no `FeatureFlags.Decode` (um `Slice[Identifier]`) e ou erra
(matando a conexão, ver P1.4) ou consome bytes errados; o Feature Flags real nunca
casa. **Fix:** `FeatureFlags` → `0x07`; adicionar tipo `UpdateTags` em `0x08` (ou
deixar 0x08 sem registro, que o framing pula limpo).

**P0.2 — `nbt.Unmarshal` de array causa `fatal error: out of memory` (não recuperável).**
`nbt/unmarshal.go:223` faz `reflect.MakeSlice(t, n, n)` com `n` vindo do payload, sem
cap. Um payload de 10 bytes declarando `TagLongArray`/`IntArray`/`ByteArray` de
`0x7FFFFFFF` tenta alocar múltiplos GB antes de ler um elemento. Confirmado
empiricamente pelo agente: **kill do processo, nem panic recuperável.** O decoder de
árvore (`decode.go`) é seguro (cap em `maxPrealloc` + grow); só o path de reflection
tem o furo. **Fix:** cap inicial em `min(n, maxPrealloc)` e crescer via
`reflect.Append` como o `list()` faz, ou validar `n*width <= bytes restantes` antes.

**P0.3 — `Slice[T].Decode` sem limite superior de count vs bytes restantes.**
`types.go` lê o count e faz `min(n, maxPrealloc)` só pra *capacity* — não rejeita um
count absurdo contra `r.Remaining()`. Hoje é incidentalmente seguro (cada elemento
consome ≥1 byte), mas vira a mesma classe da P0.2 no dia que existir um `Field` de
largura-zero. **Fix:** cross-check `n` contra `r.Remaining()` (limite de bytes por
elemento mínimo) antes de pré-alocar. Mesmo raciocínio no `bitset`.

**P0.4 — `resolveShadows` mantém os dois campos de mesmo nome/profundidade.**
`nbt/fields.go:87`: quando dois campos embedded expõem o mesmo nome na mesma
profundidade, o `encoding/json` os descarta como ambíguos; aqui os dois são mantidos
→ Marshal escreve a chave duas vezes (compound com chave duplicada = NBT inválido) e
Unmarshal popula só um, silenciosamente. **Fix:** quando >1 campo divide a
profundidade vencedora, descartar todos.

**P0.5 — VarInt overlong aceito e truncado sem erro.**
`var.go`: `FF FF FF FF 7F` decodifica como `-1` sem erro — encoding não-canônico
passa. Baixa severidade de segurança mas é correção de spec e superfície de
fingerprint. **Fix:** rejeitar quando exceder o nº máximo de bytes/bits do tipo.

**P0.6 — `BitSet`/`FixedBitSet` com índice negativo corrompe bit 63/7.**
`bitset.go`: o guard só pega `word < 0`, não `i` em `(-64, 0)`, então um índice
negativo escreve no bit errado em vez de rejeitar. **Fix:** validar `i >= 0` na
entrada de `Set`/`Get`.

---

## P1 — Robustez de protocolo e input hostil

**P1.1 — Pacotes de Configuration que o vanilla manda e não estão tratados.**
Faltam, clientbound: Plugin Message `0x00` (`minecraft:brand`), Resource Pack `0x06`,
Update Tags `0x08`. Unregistered é tolerado (pulado no framing), **mas**: servidor com
resource-pack obrigatório dá kick porque o cliente nunca responde; e o `ConfigPing`
(`0x04`) está registrado **sem handler**, então o Pong nunca é enviado → bot trava num
servidor que faz ping na fase de config. **Fix:** registrar+tratar Plugin Message (eco
do brand), Resource Pack (accept/decline) e ligar `ConfigPing → ConfigPong` no
`installConfiguration` (`session.go:60`).

**P1.2 — `Decode` não checa consumo total do payload.**
`protocol.go:104`: depois de `packet.Decode`, bytes sobrando são ignorados
(`r.Remaining()` não checado). Pacote mal-identificado "decodifica com sucesso" com
lixo — é exatamente o que deixaria a P0.1 passar batido em vez de estourar. **Fix:**
erro se `r.Remaining() != 0` após decode. (Isso vira um detector barato de bugs de ID.)

**P1.3 — Handler síncrono + write bloqueante = deadlock sob backpressure.**
`client.go:73-82`: o loop de leitura chama handlers inline, e handlers chamam
`c.Send` (um `net.Conn.Write` bloqueante). Se o peer para de drenar e nosso buffer de
socket enche, o handler trava no write, o loop não lê, o peer não lê nosso write —
deadlock clássico. Não há goroutine de escrita nem fila. **Fix:** desacoplar a escrita
numa goroutine escritora com fila bufferizada; coordenar shutdown com
`errgroup`/`context` (padrão de concorrência estruturada — reader e writer como dois
membros do grupo, primeiro erro cancela o ctx). Isso também ordena a troca de
threshold (P1.7) em relação aos frames.

**P1.4 — Um erro de decode derruba a sessão inteira.**
`client.go:86-89`: `receive` propaga qualquer erro de decode direto pra cima, então um
único pacote malformado/mal-identificado encerra tudo. **Fix:** decidir política —
tolerar erro por-pacote (log + skip) vs. fatal. Combinar com P1.2.

**P1.5 — `nbt` não suporta campo `int` puro.**
`marshal.go`: só `Int8/16/32/64`; `N int` falha com `nbt: unsupported type int`. É o
tipo inteiro mais comum de Go — usuário bate na hora. **Fix:** suportar `reflect.Int`/
`Uint*` (mapear `int`→Long, ou erro claro "use inteiro dimensionado") e documentar o
contrato signed-only.

**P1.6 — `nbt.encodeString` trunca >65535 bytes silenciosamente.**
`encode.go:92`: `PutUint16(uint16(len))` faz wrap-around → NBT corrompido sem erro.
**Fix:** erro quando o comprimento passa de `MaxUint16`.

**P1.7 — Encode de compound/map em ordem aleatória → output não-determinístico.**
`encode.go:78` / `marshal.go:61` iteram map do Go em ordem randômica; bytes diferentes
a cada run quebram comparação por bytes em teste, hashing, cache e fixtures. **Fix:**
ordenar chaves antes de emitir (ao menos como opção).

**P1.8 — Sem validação de range no encode.**
`String.Append` não valida `MaxStringLen` (só o Decode valida); `Position` mascara
fora de faixa em silêncio; `Identifier` não tem `Valid()`. **Fix:** validar no encode e
falhar cedo, simétrico ao decode.

**P1.9 — Compressão "desligada" detectada só como `-1` exato.**
`conn.go:94,115`: spec diz threshold `< 0` desliga; qualquer negativo ≠ -1 deixa
leitura esperando prefixo de data-length e escrita comprimindo → desync. Vanilla só
manda -1 ou positivo (latente). **Fix:** tratar `threshold < 0` nos dois lados.

**P1.10 — mUTF-8: bytes de continuação não validados.**
`nbt/mutf8.go:44-55`: máscaras sem checar o prefixo `10xxxxxx`; overlong/surrogate
aceitos. **Fix:** validar `b&0xC0 == 0x80` e rejeitar overlong.

---

## P2 — Design de API e consistência

**P2.1 — `Bind[T]` e `On[P]` com convenções de generic opostas + panic em uso plausível.**
`On[*SetCompression]` quer o tipo **ponteiro**; `Bind[Handshake]` quer o tipo **valor**
(`any(new(T)).(Packet)`). O `Bind[*Handshake]` simétrico compila e **dá panic no
registro** (`**Handshake` não é `Packet`). Só o `On` é checado em compile-time.
NOTA IMPORTANTE: a correção "óbvia" dos agentes é o `[T any, PT interface{*T; Packet}]`
— **exatamente o padrão de dois type-params que você rejeitou.** Então as opções que
respeitam sua decisão são: (a) guard em runtime no `Bind` que valida a asserção e dá
erro claro em vez de panic cru; (b) documentar o contrato (a assimetria é intencional,
como registramos). Não vou empurrar o two-type-param — é sua chamada.

**P2.2 — `Bind` sobrescreve silenciosamente em colisão `(state,dir,id)`.**
`protocol.go:80`: ID duplicado por copy-paste dropa um pacote sem diagnóstico. **Fix:**
detectar chave duplicada e dar panic/erro no registro. Barato e pega bug de tabela cedo.

**P2.3 — `On` chama `ID()` em ponteiro nil.**
`client.go:53-58`: `var prototype P` com `P = *Set…` é nil; `prototype.ID()` só funciona
porque todo `ID()` ignora o receiver (contrato não-forçado, não-documentado). No dia que
um `ID()` derivar de estado, quebra. **Fix:** ou alocar valor real como o `Bind` faz, ou
documentar "ID() deve ser independente do receiver".

**P2.4 — Mapa `handlers` sem sincronização.**
`client.go:18,64,111`: `On` escreve, `Run`→`dispatch` lê. Todo uso atual registra antes
do `Run` (latente), mas nada força — `go test -race` acusaria se um handler registrar
outro. **Fix:** congelar handlers antes do `Run` ou documentar o contrato
"register-before-Run".

**P2.5 — `Slice[T Field]`/`Option[T Field]` recuperam `FieldPtr` por asserção em runtime.**
Generic vazado: sem safety em compile-time, custo de type-assert por elemento no loop.
Mesma tensão da P2.1 — a correção canônica é o two-type-param que você rejeitou. Alt.:
manter como está e aceitar o custo (é pequeno), documentando a decisão.

**P2.6 — mojang: timeout por-request descartado; sem refresh token.**
`microsoft.go:99-102`: o `loginTimeout` de 10s é substituído pelo deadline do ctx (usa
um OU outro, não o menor). **Fix:** `min(ctxDeadline, now+loginTimeout)`. Além disso, o
`TokenSet.RefreshToken` e as expiries são jogados fora — a `Session` não tem
`RefreshToken`/`ExpiresAt`, então não dá pra renovar sem refazer o device-code inteiro.
**Fix:** propagar refresh token e expiry pra `Session`.

**P2.7 — Erros não são `errors.Is`-áveis; prefixo inconsistente.**
Tudo é string: `gocraft:` / `auth:` (mojang usa `auth`, não o nome do pacote) / `v765:`.
Sem sentinelas pra kick/disconnect, unknown packet, frame-too-large. `session.go:101`
embrulha o motivo do kick numa string, então o caller não detecta disconnect
programaticamente. **Fix:** exportar `ErrFrameTooLarge`, um tipo `DisconnectError{Reason}`,
etc.; padronizar o prefixo pro nome do pacote (`mojang:` em vez de `auth:`). (Respeita
sua regra de "sem sentinel errors em var" — aqui são *tipos de erro exportados* pra
API pública, não constantes internas de mensagem.)

**P2.8 — mojang não é testável sem bater no serviço real.**
Sem injeção de `Client`/base-URL no Microsoft (Yggdrasil/Mojang/Xbox já têm parcial). O
teste do estilo integração fica (é sua preferência), mas dá pra ter fakes `httptest`
pro fluxo device-code / `slow_down` / XSTS sem quebrar a regra de "teste de API externa
é integração" — os fakes cobrem a *máquina de estado*, não o serviço.

---

## P3 — Envelope do projeto (publicável / testado nas bordas / storefront)

**P3.1 — Nenhum módulo dá `go get`; deps entre irmãos só resolvem via `go.work`.**
O root importa `nbt` mas o `go.mod` não tem `require`; `codec/v765/go.mod` não requer
core nem mojang apesar de importar os dois. `GOWORK=off go build ./codec/v765` falha. Com
**zero git tags**, nada é consumível externamente. **Fix:** ou (a) colapsar pra um módulo
só com pacotes, ou (b) `require`+`replace` em cada go.mod + tags semver por-módulo.
Ver P3.4.

**P3.2 — `go test ./...` do root testa só 1 dos 5 módulos.**
`./...` não cruza fronteira de módulo no workspace — reporta `ok .../go-craft` e mais
nada. Você está sub-testando sem perceber. **Fix:** um Makefile/script que faz
`for m in . nbt mojang codec/v765 cli; do (cd $m && go test -race -cover ./...); done`.

**P3.3 — Sem CI, Makefile, lint config, LICENSE, README, `.gitignore`.**
Pra um portfólio explicitamente medido contra go-mc, a ausência de README+CI é a
primeira coisa que um revisor nota. **Fix mínimo:** README com exemplo de ping/join
rodável, LICENSE, e um GitHub Actions rodando o loop de teste por-módulo + `go vet` +
`golangci-lint` + `-race`.

**P3.4 — O split de 5 módulos não se justifica pro tamanho e adiciona atrito.**
~40 arquivos, uma versão de protocolo. O `codec/v765` importa `gocraft` em *todo*
arquivo — não está desacoplado do core, então virar módulo próprio não compra nada e
custa go.mod + sync de versão + a armadilha do P3.2. **Recomendação:** colapsar pra um
módulo só com pacotes (`/nbt`, `/mojang`, `/codec/v765`, `/cmd/cli`). Se quiser manter
algum split, mantenha só o `mojang` (único com dep externa — fasthttp — e história
standalone plausível), e só depois que tagging existir. (Isso reconsidera o go.work que
montamos; era boa prática de estudo, mas os custos hoje > benefícios.)

**P3.5 — Zero benchmark, apesar do pitch de performance.**
`grep "func Benchmark"` → nada. O design é visivelmente perf-consciente (append-based,
`maxPrealloc`, `bufio`, atomic) mas não há prova de que bate go-mc. **Fix:** `Benchmark`
pra VarInt encode/decode, round-trip de pacote completo e decode de NBT; publicar os
números vs go-mc no README — esse é o payoff de portfólio.

**P3.6 — Zero fuzzing sobre input de rede/NBT.**
`grep "func Fuzz"` → nada. Os decoders parseiam bytes hostis — alvo exato de
`go test -fuzz`. A P0.2 (OOM) teria sido pega por fuzzing trivial. **Fix:**
`FuzzReadFrame`, `FuzzVarInt` (oracle de round-trip), `FuzzDecodeNBT`, `FuzzUnmarshal`,
e um vetor canônico (`bigtest.nbt`) que toda lib NBT séria valida. Diferenciador barato
e concreto sobre go-mc. (Nota: você tinha rejeitado Fuzz "não fica bonito" no wire — mas
aqui, contra input adversário, é a ferramenta certa; dá pra manter os arquivos de fuzz
isolados num `fuzz_test.go` por pacote pra não poluir os testes de propriedade.)

**P3.7 — `session.go` (máquina de join), `mojang` OAuth e `cli` com ~0% de cobertura.**
`session.go` 0% — Login→Config→Play, keep-alive echo, teleport confirm, tudo não
testado; é onde a P0.1 vive (um teste "Update Tags em 0x08" pegava na hora). mojang 2.8%
— toda a cadeia Microsoft/Xbox/XSTS (~600 linhas, a razão de existir do módulo) sem
teste. cli 0%, nenhum `*_test.go`. **Fix:** par de `net.Conn` em memória alimentando uma
sequência de servidor canned pro session; fakes httptest pro mojang; cobra `SetArgs` +
fake server pro cli.

**P3.8 — Zero doc comment; alguns contratos não-óbvios precisam de godoc.**
Sua preferência de "sem comentários" é ok pra internals, mas pra uma *lib* o godoc é a
vitrine. Conjunto mínimo de alto valor: doc de pacote por módulo; o split `Field`/
`FieldPtr` e o modelo de erro sticky do `Reader` (erro latcha, caller checa `Err()`); o
contrato de panic de `Bind`/`On`; o sentinela `SetThreshold(-1)`; e o fato de
`NewProtocol`/`NewClient` serem obrigatórios (zero-value dá panic no mapa nil).

---

## P4 — Baixa prioridade (anotado, não urgente)

- **nbt:** unknown field é pulado decodificando+alocando a subárvore inteira e jogando
  fora (`unmarshal.go:89`) — dá pra ter um `skip(tag)` que só avança `off`. Hot path.
- **nbt:** toda string (incluindo toda chave de compound) passa por `utf16.Decode` — 2-3
  allocs por chave; fast-path ASCII resolve.
- **nbt:** `Decode` ignora bytes após o compound raiz (aceita lixo no fim); lista deriva
  o tag do `Items[0]` sem checar homogeneidade; `Encode` dá panic em `Tag` nil num
  compound.
- **core:** `nbt.go` (adapter) alcança `r.buf`/`r.off` diretos — acoplamento ao interno
  do Reader; `Identifier` wire vs `String()` divergem; `Rest()` faz aliasing do buffer.
- **codec:** `EncryptionBegin.Decode` (`login.go:34`) lê só `ServerID` e faz `r.Rest()`,
  descartando chave pública e verify token — online mode é stub (ok), mas o tipo engana.
- **codec:** `Join` usa `context.Background()` pro auth (ignora ctx/timeout do caller) e
  engole falha de decode do UUID (fica zero sem sinal).
- **status/bitset:** `statusRequestID` (0x00) reusado pra validar ID de *resposta* (lê
  como copy-paste); `BitSet.Set` cresce mas `FixedBitSet.Set` faz no-op fora de faixa.
- **testes:** `t.Run`/table tests quase não usados (35 funcs core, 27 nbt, a maioria
  flat) — falhas não localizam.

---

## Ordem de ataque sugerida

1. **P0 inteiro** — são poucos e são bugs de verdade (um derruba o processo, um quebra
   join real). Rápido e alto valor.
2. **P1.1–P1.4** — completar Configuration + goroutine escritora com errgroup + política
   de erro por-pacote. Isso destrava joins robustos e é pré-requisito real pro world
   model (a fase Play depende de um loop que não trava).
3. **P3.2 + P3.6** — script de teste multi-módulo + fuzzing. Baratos e teriam pego P0.2.
4. **P3.4** — decidir mono-módulo vs. manter split (recomendo colapsar) *antes* de
   crescer a base de código com o world model.
5. **P2 + resto de P1/P3** conforme for tocando cada área.
6. Só então **world/block/entity model**, já sobre um loop de conexão sólido.

Confirmado como **não-problema** (pra calibrar): layouts de JoinGame/SyncPlayerPosition/
LoginStart/LoginSuccess e os IDs de play/login (0x29, 0x3E, 0x24/0x15, 0x1B) batem com
1.20.4; o skip de pacote desconhecido é desync-safe (o framing consome o frame todo).
