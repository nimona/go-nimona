package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/profile"
	ishell "gopkg.in/abiosoft/ishell.v2"

	"nimona.io/go/api"
	"nimona.io/go/base58"
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
	"nimona.io/go/dht"
	"nimona.io/go/net"
	"nimona.io/go/peers"
	"nimona.io/go/storage"
	"nimona.io/go/telemetry"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"

	bootstrapPeerInfos = []string{
		// local
		// "AqApsfMu1VFWK1uVct5C8Vm1vBa91vBiyQHwrXpfNv78HZhLH5mzaFWDMHvBDkcBMKFZYQUyW9cqRA5srxzr4oFTvmd1gMfWCdwqS6kcLfjYPqdNtu5xnnWkWcjan7DcvFQrzcikPQ1y8XWudTAooUDNyGm2vTfY8HuyfngsmJWCkUoA8Br2WBcWaZTaj8TPSNznHUUAA1qDp47ZBt4VY7kgctA3f6QpLykQG4no4gDJSZpYRafP1hPKBegCKfhRavm22MrAZ7Si5xuXc1voVMTMstnSgK72cFqPaS6YWFP8Ayggoh2oGmfAPjbHAyQXvqdf92SuBzJMjAXd6Re238bL1AfuKP6PS7SyAYExLuMcfJzrDL6ndAz9wv49eXWzveCxmsh9zqhdzUX2JWTnNdyM5PFTJ2Uxn8xT9ipiZ7tGT6WAVD9NaWUDU6sZYZNpZywagwvCu4QZLi1tdGv3zB2bbcD15dZ6gWDQbHuVFHtDz6H3WkMPdmgX6vUHucvv2QBEMLRsAERQcDUKoFJpandb4ZPaTFt65Bga8P4973hTJBkDCButEuF1k4W2Jof9r6",
		// andromeda.nimona.io
		"Bm4FgCNTzLnaa6qtF6NoAhvpUNa7Q64D1sVKwvvAWPQPaWV2pSUDhbQnLcvX2Ur4NDJJDm1ZGwJuumUrZw9TqurC2GSsK2X4qLddZKaWLqPh823DnPN78r6EEKyJZAqXBSX9p1mKZbYqfAdUhjJczaYGy65jPb7vj5UCBVTqCUquB6q83hHJbDk2K1V6JGGSFqQzYGBE6M5Aw9ipVNdmSHZ8tY8isjV9tS5eV84g83LPBYVwahzVWLiqPYvo1KaTv9jAkqZyCo6WGVzLTckXTPP1MqwuAK8r9S8MecXnGqHypPj71QTpgiAHfBx4qWogZDgkPYarVpDuWXnuL9vGj5oEh5xab5uyMhVie7TXx2hxyttpkNHDNYWh16uADMhNedAwGKnQrWQnEH7uCBZp5wewHwbfzjN4FtgafdqpbnJPTiGyeGipWyBHhgWndcemEaiypwqcuciAQ8TSoxKA5B4JfKQ6q58yn465arepfxE5HtjmzAztjYSSKcz8SKXTCbPcLUjmpqRt4bCFXJ7UHo7Kxi5ETm324EmBRsdR5hjL5nLcxgRfPqvAjL39JrAtCd9hiXyenmzVSGNNJ",
		// borealis.nimona.io
		// "3SPVFKSFizVuTz4NaZ2mSmAUXDbEgfkm99c6ftUBTvmTfFGX5DwH35NvYgsWrfCxPAC5nv7T4BdF985FLG2JHdGfen51veJ4G44mwN7gA4Ltz9Ai3bLu7VZcMSnxAPTDxazACiPeWuDRdBc9AQzZMHoaDRjbAHDGXfBDDBBVJvcEAJW9EKrsFvT7NNZ5aHQZeYGx2tB42E6JwbxKJqKJHFtd227TWFH2sBzS5YymkHSUXhyC3SVN9rxWJsCdrdqg8sdTxvG45yVdpyuwZBavNR1Gd1wVJo64mdZtYcwBhDt3xABSAK9mXTkuNfgfTmJ1certQHirrXjuvfzBMzk5JtVJFai3dJJ49y8y9ZwArzvMN1SBY43miBixrW4N5fRBX4G4429GPEBLQf4V7KoKUCnVooC51Vm5Bm3cHYfRfKoFMfUmZrdBtUAB5HUAuUvjHCGrgyJwzBrpnhQByxYB2rmSJhhdzGqp21fzYCvTr8YFMF5Jt34jKR7vgsnF82QhhtXXfBm67fPRxzxRQL3VAP5bPZTSfechEaYB6HGqJMuGFLy3n5adQz3cFQGy4kmAjiwJ13GemnBMcGTx",
		// cassiopeia.nimona.io
		// "pVTMn1Bk86LBB6nmBCjCmqeYXpD9EuKQ6ybrCgKmaajx1SRFxUdAsYrsXHRNrXzF3iPKjWg2y1FyQXVZ6tmLosxTAZNkX7eyc1o69Pyp9TmS4oNu5LigF6gfLKvEYxG7EYC3dNFYnWL71ZkGSmWDVZR711fxUmgXGVeL6DCjT3mWc8V8MYqepfsN8tZDbX4ojskEpg68LYiiSwhcJTpNxCruBgn9ujV1EQPFbezYHjgf3sc8ELKo5B9U3qaB7MFSFm3YSWMZZyfZ6e4sCyj84ksw14jb4cxGE1PQKNMgMD7etTXnbzfVzZN6vG4mxh8qZTUYir9wCHVHh5K3WpXvSkzDDUxn75LzYqbHCAasejicJSbUx5V83JMyBCQLYBxSPJ2TZMHcCKUcmsTp84LDfSLDtfTfajgoaX3xdutwZeEcXxAqbeKzQzEENfAXy4u2cLJsDQE5mtoRJPAiyvCRToUAbQUdiE1yNgM4mpV6HM5fuuVShvq7ng3vsEy43HQ13v4CrUc7JPLW4wkUfZ8KZByu54Qz8xnxUCCrqW1UNW2toSj7CurQnF7uU4zMeTN8NVdiW7y8762zERYLGe",
		// draco.nimona.io
		// "2eNMjRLSSQkXHPv9Sih7b1k3vNVF4BFHPSnYBbJYEBnkAXhRe37FmiitKBHFycnnozHkjXBmou9UfAXitY8e5AtS1teeitZt8PPMjDtbtexjtxvox7KXcPXFoMPtdYjZFqRKm6tTzXN2fgUMrNvVWEc7gB6PoFow9Y2tqBMyMt3UkiXJea9pNytPXESs8hXMt4YJ8yWS6CHv75RBhh6HkRb3SXk9pLZp2T8iQR95EUDgyVxKJCnQ5T4EZWrJ66nC8EonNMnvYY22AupTPytQ5yKMdx3JFUY3LguBPD85Cz4UgGcoT9PDu93uZZmaGGZCPCoSVHq2T5J873AYxZoMwvuVs6q5TJJq2rChmQQuEjsxpw35EZQ46LeYN89jRUifTVBGRwG4iCGMkk2jZ8izzt39HC3Zk4UE8xVYXKjWg7vwKdZbzemMrA3qwnqLKtEGPTPgF5LtuG6ka9W3ZH3YA1BHpz22oJVXe2eiqLbgreq7Crwk7wb3LMjpbaPKECp3aeGkhFL8Ywcf3avVDGbb4h7ByMJfp1yj2MugozooZryME8FdJdo2NAa3Bd8XA7M4JghV8WdLpNHL",
		// eridanus.nimona.io
		// "3SPVFKSFizVuTz4NaZ2mSmAUXDbEgfkm99c6ftYYpUHnN7ymifRoZ39AYxKE3d9UTfP3hRJyEwtB8Y5G3TKagFF5tfmpp9H65XG85ZeXw1hgg6G5ESgz46yhJh4CDLKU92fhd989Q9bzc2et4iaWbudP81ix7iLQZCD4kHWp3D9UboXAxy78xXVVncditME8ZuaiRFGhufkPXkR99gnfixfe5JyZiKSAwtPWPHsydPqzcazZbFeXsLqF2CaxHdf7Tv6CHiRaAow65ZbJh43uQUa7QeDfr5qgcvVHQwXSQVCQaEtKx7CdVuLX1ofYAbUUE36nPU3pXb9ncrhXLWMDdYcnVreyM1j178fnoqcQovdVz9CYQjq9ZNhW7Gg74wDggC54rs3tMzCKDDGafqdFTwR2uY9ewU8XFPcExxqzxPAWbwn5taAPmeiAwLrWRxE8rt94wjKUAvKFU9L2mXgkcnpHpaDvxsrapaYhgDGB2CDoCTQt6dBRgNjHkW9QJnhmqBscZKzq6TaTLLPVA8S74U5jzgxDRBHrepBW8kKnoVdLC6ZJLBab7mmuvCYJQAss7myhXZfwhN4MQkeA",
		// fornax.nimona.io
		// "8FwGWEQoH9q6fN9aGX2458H1tZqf32E5UnfDbPpp51z1szWTM9S4mY8pmwqxi51HhQ4hU1831GKN1QMQnB4KQKyNt8aUbTwZyfw3HwaWirEhKaeaj5Kgkq6WtCN5NwJz2dVyLy7ncihnSWpCvYBmmm9hnDe5N1F6Lv96VEwtpsa4TWy8o13wpLC5QoTtrJ968uAv9cG8PuiqRxeaptg7gBpKAnLufS7uv8kpK5pMNU7k3SB3huQUK2xp9zPtFQYi9tp7wBsSf9koa1ysPGTvFcEC4EPfjFzZWBoWAejC1baF9hfDEmHVrnusAGu9nKuvYr49gbPE6rvGhFmHtEgRW1QQTcpgohZca8GMShy7LZHMvMq6U4qD5oAVobiKTE1uEHvdvPipcxPLf8nYkyGvQ2bHpxpowNJgvwt1QMVA2bXgBS3y21H5usTcB5GUQHuSKhHgSxhnVKfgGi6L2v6dmtg68uTfcZiAsxMVXYHSx7SzBLtrXDjbinqLFcv4c1TiTbERynrcRD6fXyEswA3gp7JMq9S7vh4F1YAUQGT1NT9tcuZxwhB8eVXrtzcqn3P3J9C2Tam8Cy6Zk",
		// gemini.nimona.io
		// "8FwGWEQoH9q6fN9aGX2458H1tZqf32E5UnfDbPtxj7MRtZZz9iNkXFVzLjo3Lc8r2KUV9pDe7L1GdcVKDpBW5ZT4Sj9qhFiVzxs3TknHEzjX1rkGsBhkhTbHW3ABcsPNpDFhqp2jpqJEgHKZQ9bRUEPvaHiNqu5TPzQfWYg1bLqB69yjuzCb5tNUa1s7dvCfEbSzFTkFy5cxao7fDd769MktxnZ4R93nuuFUbU3FGUN5H8PE5sAt4HhaeoxcG5VgufKBEGdFUpPgn7vkWD6ACEKsabq4xqVaPTCy4bZURutCzpLKRYPjkCnEiyCoGiVQnZCR6mL7b7rutb6wCSWetCriPfq4evuPy1jP7XPNMQxHRsj3QekmZtJMMkrbtFLCbBuuoPPRbVZWTVLZ7HonzNvLBipWseqbJBHiq8j9ZVc7D8srEwG2nCrn17JyFhAcagXUB7XAU6GmsxoUBkDxvBsKzkwamXypAMqRZXqfULWDh2hnJy9wzywppHKeQqCLT3g88fQCw6sovZKcHZUDFw8w8jR6VEs7z8No2HJafkkpWqau5x27PsYD63J5kj2bY5FV9GtJn7WyC",
		// hydra.nimona.io
		// "2eNMjRLSSQkXHPv9Sih7b1k3vNVF4BFHPSnYBbNUm7VwjLDTBnhoqsvPrYZDwyxpmvwHspPUGRq3kHaFmLYVnQxN31dkMfNmy9YRGt7q2wdryoj6ipX7AGKLecsqxazc8jkMih7i9B5itdWEzS2ytnuoMBYZu1hVL1an66S13eiWk9jRjQReuJRUGM9GNwUKFKLewVNXJQnydjdV4YwLnN64FsdWs9e7AExB3c57dfZej1qtsqJVJXhbVv9rQv32duERz5qrgDiENV37Qe7YrupvkmtmDiPzavUANkUiNA8UhXCfXwihdN81QPYbAikKUV4YcfxEjyaMAWgBs7ty7gxWG7nST9hvi3nNuNVu4hHANcNcDx1Q5qe7WkGQt4EHedjKGXdqQ7kWLLuVfooHKi7iYRzDSeHmitGFkAJHiaug4t4xBvh76wQZmPqW1ne319JMZoSvr3e7btSiyrQacy5ZGRKuYXfMphHeFghuotJnPuGNtt64PW2aMGbzMwLVJuA1oauWEZ8nm6UZncbJiwLsvcKRUeUDb4fZoWFg4xVeXGGSZ6CYXDPnxZMTJxeTrFXXjXD2FmXt",
		// indus.nimona.io
		// "2eNMjRLSSQkXHPv9Sih7b1k3vNVF4BFHPSnYBbPR4tZ4txJRVCJV28vftLW2ppRTQGVQ6Qg8uMa7bMWV6EqzqXJETHRSqVEQBrcBnz1vFPLPyrkKnjduCBRe2qfQr4Y3QYAkS2MB7WLbpaXZ2LeF5L5KV1vgzeZRWW1EY7KTjqxhx7z7gMrEAsRCA1kG6Qc2AxB1XwWu73iMm4TwCsg8cecvw3xGndcH9iunek7HXk217fu3V5HUEBY1VzQoGgXkuuYVTYD3FzCRWCJAWevEsVVMPhc7Rqqveuns8wUf6CkJhW5Mrm9xFXFgH81xTtTCFGAZ6D2pXURVr2Cu62XFA4diFzygW21ivprVyzTU7pU8m9XJ5dfgmqBikxgaZ4GD5pHJhYCGjhtKWvg3qWL96tQo6FPdT8H9CjCXFzF8utLVRV78J1ordD68ZiJrELvopRujRgfmJR8MvvHWG8AKwXyQCFNhKrSrdBx5i7ne2jAt3FLuQRy18A1KhYJ812TxH42hvhGbPoYr1j2EsZ5yyudThQ5vLHGqtbTH3NMMTDq7uVSRuQbdqDKaXZTPt8pLhEecdGKCHmvE",
		// lacerta.nimona.io
		// "Z2vXS2C3GHyHzLLqqVVeZD7gvbwksvQtG9EwnSD9BqA2yiTDXpAe7FYrDV3baSWKKWQ5J5qBKsyGstuK9WttExNXQNm7HoLb2XWd18Afcb3FksLjp9Sa2G8TrgaR8JQMyg5EaucB3mfcJDXZq9UWF6StqZ5fuGPKAoN2psPgkRFwiGh4pgtFgXB7xZp3TpyUJcjfAvsmEQRpLmLFUFHkePB3ZCjKUiartaP4SnS7E6b9xzsBVHwaSPhU2pr5pgia4Zhvz84P4zoL5eGftLtz6YvVT8peeroyhGEQxUZ4TCWnde7T4cpJpWbwn3dbRofFhvcPgB9DCDYwskvbt8WxJvK1Szuqt3XA5HeE4uyzYKjGB4eJqM22kVUdVUW6kiKeby7PC5xgGRsaTuCLvss1VZgu2Pp1UbbF7wH9SEuDLLx1Tbv8nXrZbUemuaMYR5gcX6A2BkkcbWEj8ZuDxFxVsr8ABecgCeZ1Mn5LtudAwXvE2sXFPRMQQ6fmGD2JaeFJWFHCWH8JktsDGuh5moEfAVEW5SJF2QjuuEgdBur64QWSmpdDADNYUgkchYH7nrGkJgrkguen3oaq3t",
		// mensa.nimona.io
		// "2eNMjRLSSQkXHPv9Sih7b1k3vNVF4BFHPSnYBbTJ4TsvUFFWAFDuv2GfL3tXtFZhZT85K7ijtzkYqsZDMn5bMuYTPGULdaPoY8VhzcQXHV5NKkiX2mFHcj2Kh48si6sjoa1trb2ZfCUFcxQhGEAS6iHCZuDXU1vmvCURgdagpNYRCmfVPaSuh98rvGnsgnmrusg8UhyHBnYcdnzhiRG6JZfuUducpHAs5uAkvW38VRZtYbqw2PKXzNZPoguSS6WjjdWZSA6rjgEdnFgfWDEk2gxycnkYDatbz2BCaQ5kq1AMeQ151iffu5HoDJM24mdXZ5YqVbJfD5i2V7SaWAMcNHHP73nSJkQiyQGPtnfNW1XqCHz7Mhbr6cRsN2o1b2wXEfcmdc9ujoi34RwkAtZNjLZo1dbyUtPL9vueojt7qGxPpKGpP9aLY98HEuxiGYBkXQF2MFYjvmNFhfAdBPDjActiqNk71vYknTYb1F4j2zXVwgnp2eS4r61sVsNUYBR1n3rHTfDagEKiEU2Sh1nFmrjJcvPyj2RouHgaNR5vSk13iF6WeqPZVbDCJ5PZ8GBBRZhsrF7g1Pmp",
		// norma.nimona.io
		// "2eNMjRLSSQkXHPv9Sih7b1k3vNVF4BFHPSnYBbUK2RNHHiHCcurP65EzQBivqXgqMcirhAEtfRR2jdBvJqXmT64juQ5pGw7xyyhaf1oC9MNw37QQ5uoN4dLHn4GTkxdTv1uViV6qwSQJVN5f6qmd3k9kwL9DdrNDXwC4ufTmSMNEa82xEf9W4rp365SFgbhN6hhUDw4SJFhXgfoqUe5A2atGwL3A8yawXKweQLAMJ85Cqt8tMZLfjcn4JyoPHwdbtpwPUSUij9s44rcxsDfqD4XYVr3D5hwonAeh9HpCTkD9eSCnMsED62gBsa4RgpTWiTys6yeMm2eeRPnFRk24ARu1XJpewPNCsvKLvKeFRee73PH5SfooPx97wpnrTsHQD1KG3BQ8GnC8WHEtekVTiNB3aYAvJPAu64kw5dWVy5NTwAqcpTBNgtYmMzbeLJReSmXuGVBrmoGPi4HhJQBiWqMqhHGVQmCTm3cXFeBntkb64kwEsYezCrqZ3omMsFeQdGo6RfLoMSpEHqZrteaPmsQLuW56d51QNxCtNM5KBmmQXzXqTf89B24xQABJdmampWQepZwEiYez",
		// orion.nimona.io
		// "2eNMjRLSSQkXHPv9Sih7b1k3vNVF4BFHPSnYBbVJRjTFzpzaMVnpdsKzYyTSm5YKiz28WueSTM3hfWoKmtrnkyJv4BXDGswuUTprcM69n35jSa2fFdvJFL63JBX41hxQBdfV5rxRnpfcuNbrRR8hved5Ad5QwYn8CEX1ZxUkQB4cbXyqoq2Q12yAvGR1ipZbc4jUKPYw4p4P8cXrerbXv9Ur3ExgPKTYkBfe6AE62SR7Mqb2fkf6TegMH4cH8MrTHteUuXq6ZqA8v4guYgdDDfNvDzjGEaNnUwC926t23PLxATVi3NAaL2wARzdWnp7KzutuPp9kykNSmyxEEpnnCeBBCeqoMKUT9KDTwtZCNyToJPq2EQZMvp9TTG5n2MEabhs8hb7eYXkV8v5wJb24e9KZhyMzTnX1NeLcFgVNQfDeqyEizZAsb36QyCdkBdBbThV2yqBVLN7F7YGSkUksMZV8EcpTdfuXKVHFKFekZGFuSJHEMJQj91ChKGyuqbVgxHnxHMsYgpja9Ps8j41jEiBcRDRyAVAR2gU9eMBwSFyLpJLq1k2U3bMQGNXCZvAZC6DQvpzGE78i",
		// pyxis.nimona.io
		// "2eNMjRLSSQkXHPv9Sih7b1k3vNVF4BFHPSnYBbWJjfztib2k1RiYbekckLURUZE4sfCb99NWBa1ngyMmxSnkXC2Mamv2LDsK8vDYTRQrdMbyF9cQU9eqRCsG9wUcHWzHvA1SNP9QX9YtDeANyt9F1NiMewRDgaNd86pZsRyEJQVxcgv1p6SyVHEcpQbYM7VxkdJJQsXqU718D2kqUZpZNgw1379dp2zhyfc6FX9EXanXWLEe6AC12y9AbTxVSGtQkrhkRVtuxUbcBTSyuD1zdu2nEwGRTQFCTtC7myLWHRhQCiy1XYKTQTZw8qRzNGpmuLRYi6xETwzw1wWu4c7iLTirKkfxLk2UJ1aDLokvhPdWtLsjdWDQdVPrLVkSNwburHJz1Sm5rsdjFowfXy5KdzBCTtSF2u2Ab8ECmBGhrf5e6NuzALXBgDdVhGyaVvoij5Z6mpp7hnKjTCG9aabeQncWiA4T5pZvB8BPjBQgpg16qP9b4eYvJQC86i2KyQjXWFCWrBsaSi2Vjo1NrMysWnQBTLV5UDE8cpeDSAjeHQvznfEugmPGr6kpYZf5TGftAD1ptEtA39NE",
		// stats.nimona.io
		"2DQbyaTKWKoPAfvq3nzBRUyAFzH7Zf6mFroa9Hr6EBKvcgDKXSxbEk4njcxmsxzsKfeHiPM45qwZ8hZp6eMY4URu1eJzwxhqQYuPYhYEXyyvXtE9eiDnXcv5Cq6GCAPf4phSp47Zz7QVzcsrZMQmKAewaQRWgbaPKU9Nymg5rPAuneMDa8eBbMnWYHkiRWCXJE5xdudvdVBteZ6BcqWh6Mbe5h8rZevmrzPxBhxQUonKDFCHApihkdWJ7wzb9umBzPMinH4GEa3mp6TNGeLUz3g7Sutk4B3D7E2dtTZxvZQVdJFXz5HrNPBdCSZxbqQMgT8Rje9pvqGx7i6kZdyiAFMWjaZUZYSShezFvQvbCxTiTL4dCgfPEqiMXMavxobq4SjFw92sEJwhPWKwf8atUU1Jw8j3UpMgUtcCdYEaSBQm21oSzEFFYn8Ucs1UDn3tRBbJqCDPQwpqRf7cR7Ujo7ApxKwQXw5nmZKfxbAqma2XGiQFySrxhE9yzPv7DK3FLdL1R6TPJPzNMXoZoWBQqfV9YCofyvWYzrieeYQdpvhjSP8QuqWypaYeEVpMGUUpSfVymgR5hQthFvDS1sGoCgn",
	}
)

func base64ToBytes(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// Hello payload
type Hello struct {
	Body      string            `json:"body"`
	Signature *crypto.Signature `json:"-"`
}

func (h *Hello) GetType() string {
	return "demo.hello"
}

func (h *Hello) GetSignature() *crypto.Signature {
	return h.Signature
}

func (h *Hello) SetSignature(s *crypto.Signature) {
	h.Signature = s
}

func (h *Hello) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (h *Hello) SetAnnotations(a map[string]interface{}) {
	// no annotations
}

func init() {
	blocks.RegisterContentType(&Hello{}, blocks.Persist())
}

func main() {
	defer profile.Start(profile.MemProfile).Stop()

	configPath := os.Getenv("NIMONA_PATH")

	if configPath == "" {
		usr, _ := user.Current()
		configPath = path.Join(usr.HomeDir, ".nimona")
	}

	if err := os.MkdirAll(configPath, 0777); err != nil {
		log.Fatal("could not create config dir", err)
	}

	reg, err := peers.NewAddressBook(configPath)
	if err != nil {
		log.Fatal("could not load key", err)
	}

	port, _ := strconv.ParseInt(os.Getenv("PORT"), 10, 32)

	bootstrapPeer := &peers.PeerInfo{}
	for _, peerInfoB58 := range bootstrapPeerInfos {
		peerInfoBytes, _ := base58.Decode(peerInfoB58)
		peerInfo, err := blocks.UnpackDecode(peerInfoBytes)
		if err != nil {
			panic(err)
		}
		if err := reg.PutPeerInfo(peerInfo.(*peers.PeerInfo)); err != nil {
			log.Fatal("could not put bootstrap peer", err)
		}
		bootstrapPeer = peerInfo.(*peers.PeerInfo)
	}

	storagePath := path.Join(configPath, "storage")

	dpr := storage.NewDiskStorage(storagePath)
	n, _ := net.NewExchange(reg, dpr)
	dht, _ := dht.NewDHT(n, reg)
	telemetry.NewTelemetry(n, reg.GetLocalPeerInfo().Key,
		bootstrapPeer.Signature.Key)

	n.RegisterDiscoverer(dht)

	n.Listen(context.Background(), fmt.Sprintf("0.0.0.0:%d", port))

	n.Handle("demo.hello", func(payload blocks.Typed) error {
		fmt.Printf("___ Got block %s\n", payload.(*Hello).Body)
		return nil
	})

	httpPort := "26880"
	if nhp := os.Getenv("HTTP_PORT"); nhp != "" {
		httpPort = nhp
	}
	httpAddress := ":" + httpPort
	api := api.New(reg, dht, dpr)
	go api.Serve(httpAddress)

	shell := ishell.New()
	shell.Printf("Nimona DHT (%s)\n", version)

	putProvider := &ishell.Cmd{
		Name:    "providers",
		Aliases: []string{"provider"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			if len(c.Args) < 1 {
				c.Println("Missing providing key")
				return
			}
			key := c.Args[0]
			ctx := context.Background()
			if err := dht.PutProviders(ctx, key); err != nil {
				c.Printf("Could not put key %s\n", key)
				c.Printf("Error: %s\n", err)
			}
		},
		Help: "announce a provided key on the dht",
	}

	getProvider := &ishell.Cmd{
		Name:    "providers",
		Aliases: []string{"provider"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			if len(c.Args) == 0 {
				c.Println("Missing key")
				return
			}
			c.ProgressBar().Indeterminate(true)
			c.ProgressBar().Start()
			key := c.Args[0]
			ctx := context.Background()
			rs, err := dht.GetProviders(ctx, key)
			c.Println("")
			if err != nil {
				c.Printf("Could not get providers for key %s\n", key)
				c.Printf("Error: %s\n", err)
			}
			providers := []string{}
			for provider := range rs {
				providers = append(providers, provider.Thumbprint())
			}
			c.Println("* " + key)
			for _, peerID := range providers {
				c.Printf("  - %s\n", peerID)
			}
			c.ProgressBar().Stop()
		},
		Help: "get peers providing a value from the dht",
	}

	getBlock := &ishell.Cmd{
		Name:    "blocks",
		Aliases: []string{"block"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			if len(c.Args) < 1 {
				c.Println("Missing block id")
				return
			}

			ctx, cf := context.WithTimeout(context.Background(), time.Second*10)
			defer cf()

			block, err := n.Get(ctx, c.Args[0])
			if err != nil {
				c.Println(err)
				return
			}

			c.Printf("Received block %#v\n", block)
		},
		Help: "get peers providing a value from the dht",
	}

	listProviders := &ishell.Cmd{
		Name:    "providers",
		Aliases: []string{"provider"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			ps, _ := dht.GetAllProviders()
			for key, vals := range ps {
				c.Println("* " + key)
				for _, val := range vals {
					c.Printf("  - %s\n", val)
				}
			}
		},
		Help: "list all providers stored in our local dht",
	}

	listPeers := &ishell.Cmd{
		Name:    "peers",
		Aliases: []string{"peer"},
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			ps, _ := reg.GetAllPeerInfo()
			for _, peer := range ps {
				c.Println("* " + peer.Thumbprint())
				c.Printf("  - addresses:\n")
				for _, address := range peer.Addresses {
					c.Printf("     - %s\n", address)
				}
			}
		},
		Help: "list all peers stored in our local dht",
	}

	listBlocks := &ishell.Cmd{
		Name: "blocks",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			blocks, err := dpr.List()
			if err != nil {
				c.Println(err)
				return
			}
			for _, block := range blocks {
				c.Printf("     - %s\n", block)
			}
		},
		Help: "list all blocks in local storage",
	}

	listLocal := &ishell.Cmd{
		Name: "local",
		Func: func(c *ishell.Context) {
			c.ShowPrompt(false)
			defer c.ShowPrompt(true)

			peer := reg.GetLocalPeerInfo()
			c.Println("* " + peer.Thumbprint())
			c.Printf("  - addresses:\n")
			for _, address := range peer.Addresses {
				c.Printf("     - %s\n", address)
			}
		},
		Help: "list protocols for local peer",
	}

	send := &ishell.Cmd{
		Name: "send",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Println("Missing peer id or block")
				return
			}
			ctx := context.Background()
			msg := strings.Join(c.Args[1:], " ")
			peer, err := reg.GetPeerInfo(c.Args[0])
			if err != nil {
				c.Println("Could not get peer")
				return
			}
			signer := reg.GetLocalPeerInfo().Key
			if err := n.Send(ctx, &Hello{Body: msg}, peer.Signature.Key, blocks.SignWith(signer)); err != nil {
				c.Println("Could not send block", err)
				return
			}
		},
		Help: "list protocols for local peer",
	}

	block := &ishell.Cmd{
		Name: "block",
		Help: "send blocks to peers",
	}

	get := &ishell.Cmd{
		Name: "get",
		Help: "get resource",
	}

	get.AddCmd(getProvider)
	get.AddCmd(getBlock)

	put := &ishell.Cmd{
		Name: "put",
		Help: "put resource",
	}

	put.AddCmd(putProvider)

	list := &ishell.Cmd{
		Name:    "list",
		Aliases: []string{"l", "ls"},
		Help:    "list cached resources",
	}

	list.AddCmd(listProviders)
	list.AddCmd(listPeers)
	list.AddCmd(listLocal)
	list.AddCmd(listBlocks)

	shell.AddCmd(block)
	shell.AddCmd(get)
	shell.AddCmd(put)
	shell.AddCmd(list)
	shell.AddCmd(send)

	// when started with "exit" as first argument, assume non-interactive execution
	if len(os.Args) > 1 && os.Args[1] == "exit" {
		shell.Process(os.Args[2:]...)
	} else {
		// start shell
		shell.Run()
		// teardown
		shell.Close()
	}
}
