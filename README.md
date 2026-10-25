🚀 Przewodnik Wdrożenia Nowego Mikroserwisu
Ten dokument opisuje standard wdrożeniowy dla nowych aplikacji uruchamianych na serwerze Proxmox LXC 102 (`192.168.0.81`) przy użyciu Dockera, Portainera, globalnej bazy PostgreSQL oraz Cloudflare Tunnel.
---
🏗️ 1. Architektura i Zasady Działania
Infrastruktura: Kontener LXC 102 z zainstalowanym silnikiem Docker oraz Portainerem.
Baza Danych: Używamy jednej globalnej bazy PostgreSQL (`global_postgres`). Każdy serwis ma całkowicie odizolowany schemat wewnątrz bazy (np. `test_schema`, `immich_schema`), co zapewnia bezpieczeństwo i porządek bez marnowania zasobów na osobne kontenery bazodanowe.
Sieć (`web-network`): Specjalna zewnętrzna sieć dockera, która spina ze sobą mikroserwisy oraz globalną bazę danych, umożliwiając im komunikację po nazwach hostów (np. `global_postgres`).
Dostęp Zewnętrzny (Cloudflare Tunnel): Ruch z Internetu trafia przez bezpieczny tunel Cloudflare bezpośrednio na nasz serwer i jest kierowany na odpowiedni port kontenera na podstawie pliku konfiguracyjnego `config.yml`.
---
🛠️ 2. Krok po kroku: Jak wdrożyć nowy serwis
Krok 1: Przygotowanie bazy danych (Opcjonalnie)
Jeśli Twój serwis potrzebuje bazy PostgreSQL:
Połącz się z kontenerem `global_postgres`.
Utwórz dedykowany schemat dla nowej aplikacji:
```sql
   CREATE SCHEMA nazwa_serwisu_schema;
   ```
W konfiguracji aplikacji wskaż połączenie do `global_postgres` z parametrem ustawiającym domyślny schemat (np. `search_path=nazwa_serwisu_schema`).
---
Krok 2: Przygotowanie kodu i `Dockerfile`
Używamy najprostszych, jednowarstwowych obrazów (bez zbędnego komplikowania z multi-stage builds), aby uniknąć błędów kompilacji.
Przykład najprostszego `Dockerfile` dla aplikacji w Go:
```dockerfile
FROM golang:1.26-alpine

WORKDIR /app

# Kopiujemy kod projektu
COPY . .

# Pobieramy zależności
RUN go mod tidy

EXPOSE 8080

# Uruchamiamy aplikację bezpośrednio
CMD ["go", "run", "main.go"]
```
---
Krok 3: Konfiguracja Stacku w Portainerze
Utwórz nowe repozytorium na GitHubie dla swojego serwisu.
W Portainerze przejdź do Stacks -> Add stack.
Połącz z repozytorium GitHub i wklej plik `docker-compose.yml` o następującej strukturze:
```yaml
version: '3.8'

services:
  app:
    image: twoje-repo/nazwa-serwisu:latest
    # Lub budowanie bezpośrednio z GitHuba przez Portainer:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: nazwa_serwisu_app
    restart: unless-stopped
    ports:
      - "8082:8080" # Mapujemy port zewnętrzny (np. 8082) na port aplikacji w kontenerze (8080)
    environment:
      - DB_HOST=global_postgres
      - DB_NAME=main_db
      - DB_SCHEMA=nazwa_serwisu_schema
      - DB_USER=twoj_user
      - DB_PASSWORD=twoje_haslo
    networks:
      - web-network

networks:
  web-network:
    external: true
```
> **Ważne:** Upewnij się, że sieć `web-network` jest zaznaczona jako `external: true`.
---
Krok 4: Spojrzenie na sieć Dockera
Aby aplikacja mogła rozmawiać z globalną bazą danych, upewnij się, że kontener bazy danych jest podpięty do tej samej sieci:
```bash
docker network connect web-network global_postgres
```
---
Krok 5: Publikacja pod domeną (Cloudflare Tunnel)
Wejdź na serwer LXC 102 i edytuj plik konfiguracyjny tunelu Cloudflare (zazwyczaj `/etc/cloudflared/config.yml`):
```bash
   nano /etc/cloudflared/config.yml
   ```
Dopisz swój nowy serwis w sekcji `ingress:` tuż nad regułą domyślną `404`:
```yaml
   ingress:
     - hostname: immich.rajmanski.com
       service: http://192.168.0.157:2283
     
     # Twój nowy serwis:
     - hostname: moj-serwis.rajmanski.com
       service: http://192.168.0.81:8082

     # Reguła domyślna (zawsze na samym końcu!):
     - service: http_status:404
   ```
Zapisz plik (`Ctrl + O`, `Enter`, `Ctrl + X`).
Zrestartuj usługę tunelu, aby wczytać zmiany:
```bash
   sudo systemctl restart cloudflared
   # lub jeśli działa jako kontener w Dockerze:
   # docker restart nazwa_kontenera_cloudflared
   ```
Utwórz rekord DNS dla nowej subdomeny w Cloudflare (wykonaj raz w konsoli):
```bash
   cloudflared tunnel route dns b6d2ae24-3760-450b-b776-752b64116e33 moj-serwis.rajmanski.com
   ```
---
🎉 Gotowe! Twój nowy serwis działa pod nową domeną, jest całkowicie odizolowany, spięty z bazą i automatycznie budowany przez Portainera.