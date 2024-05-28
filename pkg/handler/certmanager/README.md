# certmanager

Managing CA and Certificates

GET /ca - Returns all CA
POST /ca - Create a new CA
GET /ca/:id - Returns information about a CA key
PUT /ca/:id/verify - Verirfy a CA
DELETE /ca/:id - Deletes a CA
TODO: Renew a CA

GET /cert - Returns all Certificates
POST /cert - Create a new Certificate
GET /cert/:id - Returns a certificate public key
DELETE /cert/:id - Deletes a certificate
TODO: Renew a certificate

CA:
Passphrase (4 to 1023 characters)
Days of validity
Subject
Country
State
Location
Organization
Organizational Unit

# generate aes encrypted private key
openssl genrsa -aes256 -out ca.key 4096

# make public cert
openssl req -x509 -new -nodes -key ca.key -sha256 -days 1826 -out ca.crt -subj '/CN=test root CA/C=DE/ST=Berlin/L=Berlin/O=test'

# cert info
openssl x509 -in ca.crt -text -noout

# verify cert
openssl verify -CAfile ca.crt ca.crt

# check a private key
openssl rsa -in ca.key -check

# comvert to pem
openssl x509 -inform der -in ca.crt -out ca.pem

# pfx
openssl pkcs12 -export -out ca.pfx -inkey ca.key -in ca.crt -certfile ca.crt