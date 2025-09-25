# Arrakis - Biblioteca de Polling Adaptativo SQS para Go

Uma biblioteca Go que implementa polling adaptativo para Amazon SQS usando algoritmos EWMA (Exponentially Weighted Moving Average). A Arrakis otimiza automaticamente os intervalos de polling baseado no volume de mensagens, reduzindo custos de API e melhorando a responsividade.

## 🚀 Características

- 🎯 **Polling Adaptativo Inteligente**: Ajusta automaticamente os intervalos de polling SQS
- 📊 **Algoritmo EWMA**: Usa média móvel exponencial para detectar padrões de volume
- 💰 **Otimização de Custos**: Reduz chamadas de API desnecessárias durante períodos ociosos
- ⚡ **Baixa Latência**: Polling frequente durante picos de tráfego
- 🛡️ **Proteção contra Picos**: Previne distorções causadas por outliers
- � **Detecção de Quedas**: Adapta-se rapidamente a reduções no volume de mensagens
- 📈 **Decay Temporal**: Reduz gradualmente a frequência durante períodos ociosos
- �️ **Altamente Configurável**: Parâmetros ajustáveis para diferentes cenários

## 📦 Instalação

```bash
go get github.com/elissonalvesilva/arrakis
```

## 🎯 Uso Básico

```go
package main

import (
    "context"
    "log"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/elissonalvesilva/arrakis/pkg/sqs"
)

func main() {
    // Carregar configuração AWS
    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Criar cliente SQS com Arrakis
    sqsClient := sqs.NewSQS(&cfg)
    
    // Ativar polling adaptativo - é aqui que a mágica acontece!
    sqsClient.EnableArrakis()
    
    // Usar normalmente - Arrakis otimiza automaticamente
    queueURL := "https://sqs.us-east-1.amazonaws.com/123456789012/minha-fila"
    messages, err := sqsClient.ReceiveMessage(context.Background(), queueURL, 10, nil)
    if err != nil {
        log.Printf("Erro ao receber mensagens: %v", err)
        return
    }
    
    log.Printf("Recebidas %d mensagens com polling adaptativo", len(messages.Messages))
}
```

## ⚙️ Configuração Avançada

```go
// Configuração personalizada para cenários de alto volume
option := sqs.WithAdaptivePolling(
    20, // idleWaitTimeSeconds - tempo de espera quando não há mensagens
    60, // visibilityTimeout - timeout de visibilidade das mensagens
    12, // lowVolumeWaitTimeSeconds - espera para baixo volume
    8,  // mediumVolumeWaitTimeSeconds - espera para volume médio
    4,  // highVolumeWaitTimeSeconds - espera para alto volume
    1,  // veryHighVolumeWaitTimeSeconds - espera para volume muito alto
    0.4, // ewmaAlpha - fator de suavização (mais responsivo)
    8,   // dropDetectionThreshold - ciclos antes de resetar EWMA
)

// Aplicar configuração personalizada
sqsClient := sqs.NewSQS(&cfg, option)
sqsClient.EnableArrakis()
```

## 📊 Como Funciona

A Arrakis classifica automaticamente o volume de mensagens em categorias e ajusta os intervalos de polling:

| Volume | Critério (EWMA) | Tempo de Espera | Cenário |
|--------|---------------|----------------|---------|
| **Idle** | = 0 mensagens | 20 segundos | Fila vazia |
| **Low** | < 2 mensagens | 15 segundos | Tráfego baixo |
| **Medium** | 2-5 mensagens | 10 segundos | Tráfego moderado |
| **High** | 5-10 mensagens | 5 segundos | Tráfego intenso |
| **Very High** | > 10 mensagens | 1 segundo | Pico de tráfego |

### Algoritmo EWMA
```
novo_valor = α × observação_atual + (1-α) × valor_anterior
```

- **α baixo (0.1-0.3)**: Mais estável, mudanças graduais
- **α alto (0.4-0.7)**: Mais responsivo, adaptação rápida
- **Recomendado**: 0.2-0.4 para a maioria dos casos

## 📈 Benefícios

### Redução de Custos
- **Até 70% menos chamadas** de API durante períodos ociosos
- Polling inteligente baseado em padrões reais de tráfego
- Prevenção de polling desnecessário

### Melhoria de Performance
- **Latência reduzida** durante picos de tráfego
- Adaptação automática a mudanças de volume
- Eliminação de configuração manual de intervalos

### Confiabilidade
- Proteção contra picos isolados
- Detecção automática de quedas de volume
- Recovery rápido após períodos ociosos

## 🏗️ Estrutura do Projeto

```
arrakis/
├── pkg/sqs/                    # API pública da biblioteca
│   ├── sqs.go                 # Cliente SQS principal
│   ├── arrakis.go             # Algoritmo de polling adaptativo
│   ├── options.go             # Configurações e opções
│   └── sqs_test.go            # Testes unitários
├── pkg/internal/infra/utils/   # Utilitários internos
├── examples/                   # Exemplos de uso
└── docs/                      # Documentação técnica
```

## 📚 Documentação

- [Documentação Técnica](docs/TECHNICAL.md) - Detalhes do algoritmo EWMA
- [Exemplos de Uso](examples/) - Casos de uso práticos
- [API Reference](docs/API.md) - Referência completa da API

## 🧪 Testes

```bash
# Executar todos os testes
make test

# Executar testes com cobertura
make test-coverage

# Executar benchmarks
make benchmark
```

## 🤝 Contribuindo

1. Fork o projeto
2. Crie sua feature branch (`git checkout -b feature/amazing-feature`)
3. Commit suas mudanças (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🙏 Inspiração

O nome "Arrakis" é uma homenagem ao planeta do universo Dune, conhecido por seus recursos valiosos e pela necessidade de otimização para sobrevivência - assim como esta biblioteca otimiza o uso de recursos SQS.