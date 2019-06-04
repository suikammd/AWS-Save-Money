Add your server public key and secret key here

1. Secret key

   ```shell
   openssl ecparam -genkey -name secp384r1 -out server.key
   ```

   

2. Public key

   ```shell
   openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
   ```