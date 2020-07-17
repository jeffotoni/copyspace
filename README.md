# copyspace
copyspace é uma ferramenta que copia arquivos para os Buckets da DigitalOcean chamados de Spaces

#### Install
```bash
$ sh -c "$(wget https://raw.githubusercontent.com/jeffotoni/s3godo/master/spaces/copyspace/v1/install.sh -O -)"
```
#### Credenciais

Precisará criar um arquivo oculdo em seu home, .dokeys nele precisa conter suas credenciais.
```bash
{
    "key":"xxxxxxxxxxxx",
    "secret":"xxxxxxxxxx",
    "endpoint":"https://sfo2.digitaloceanspaces.com",
    "region":"us-east-1",
    "bucket":"your-bucket"
}
```

Feito isto agora ficou fácil, basta executar o copyspace
```bash
$ copyspace --file=your-file.pdf --acl=public --bucket=your-bucket
```
Copyspace também copia recursivamente
```bash
$ copyspace --file=/home/user/projetos --acl=public --bucket=your-bucket
```
