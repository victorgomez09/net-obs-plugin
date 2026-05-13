# --- ETAPA 1: Compilación (Builder) ---
FROM golang:1.26-alpine AS builder

# 1. Instalamos dependencias del sistema para compilación de C (necesario para libpcap/gopacket)
RUN apk add --no-cache \
    gcc \
    musl-dev \
    libpcap-dev \
    linux-headers

WORKDIR /app

# 2. Gestión de módulos (Capa cacheable)
# Copiamos solo los archivos de dependencias primero
COPY go.mod go.sum ./
RUN go mod download -x

# 3. Copiamos el resto del código fuente
# IMPORTANTE: Asegúrate de que los archivos .go estén en la raíz del proyecto
COPY . .

# 4. Verificación de archivos (Para depurar el error "no Go files")
RUN ls -la *.go

# 5. Compilación del binario
# CGO_ENABLED=1 es obligatorio para gopacket
# GOARCH=amd64 (Cámbialo a arm64 si vas a las Raspberry)
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -o net-oracle .

# --- ETAPA 2: Imagen Final (Runtime) ---
FROM alpine:latest

# 1. Instalamos la librería de ejecución para pcap
RUN apk add --no-cache libpcap

WORKDIR /root/

# 2. Traemos el binario desde la etapa de compilación
COPY --from=builder /app/net-oracle .

# 3. Permisos de ejecución (por si acaso)
RUN chmod +x net-oracle

# 4. Comando de inicio
ENTRYPOINT ["./net-oracle"]