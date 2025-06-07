import {CoinBalance, getFullnodeUrl, SuiClient} from "@mysten/sui/client"
import { Ed25519Keypair} from "@mysten/sui/keypairs/ed25519"
import {getFaucetHost, requestSuiFromFaucetV2} from "@mysten/sui/faucet"

import {Transaction} from "@mysten/sui/transactions"
import dotenv from "dotenv"
import fs from "fs"
dotenv.config();


interface Transfer {
    to:string, 
    amount: number
}


const wallets = JSON.parse(fs.readFileSync('wallets.json', 'utf-8'))


const suiClient = new SuiClient({
    url: getFullnodeUrl("testnet")
})

const transferFromWallet = async (wallet: {address: string, privateKey: string}) => {
    try {

        const pair  = Ed25519Keypair.fromSecretKey(wallet.privateKey)

       
        const recipient = process.env.ADDRESS!
        const amount = 10000000

        const tx = new Transaction();
        tx.setSender(pair.getPublicKey().toSuiAddress())


        const [coin] = tx.splitCoins(tx.gas, [amount])
        tx.transferObjects([coin], recipient)


        const result =  pair.signAndExecuteTransaction({
            client: suiClient,
            transaction: tx
        })

        // await suiClient.waitForTransaction({digest: (await result).digest})
        console.log("Tx executed", (await result).digest)

    }catch(error){
        
        console.error(`Transaction failed for wallet ${wallet.address}:`, error);
    }
}


(async()=>{
    await Promise.all(wallets.map(transferFromWallet))
})()