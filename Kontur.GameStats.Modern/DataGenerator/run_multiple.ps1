$args = "--put", "1000", "100", "50"

Clear-Host;

for ($i = 1; $i -le 1; $i++) {
    Start-Process -NoNewWindow "bin/Debug/net10.0/DataGenerator" -ArgumentList $args
}
