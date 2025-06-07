import { getFullnodeUrl, SuiClient} from "@mysten/sui/client"
import { Ed25519Keypair} from "@mysten/sui/keypairs/ed25519"

import {Transaction} from "@mysten/sui/transactions"
import dotenv from "dotenv"
dotenv.config();



(async()=>{
const suiClient = new SuiClient({
    url: getFullnodeUrl("testnet")
})


const keyPair = Ed25519Keypair.fromSecretKey(process.env.PRIVATE_KEY!)

const tx = new Transaction();
tx.setSender(keyPair.getPublicKey().toSuiAddress())
const amount = 100000000

const coin = tx.splitCoins(tx.gas, [0])

tx.moveCall({
    target :"0x6706b812a8ae5512deb9723e128a55e2191a8e533c221c4a725d9dc423c924c8::nft_with_fees::mint_nft",
    arguments: [
        tx.object("0xa57661d20a501820f4c69c533cfcf7d970646242e41f715091ef29021e9c8d57"), // Collection name 
        tx.pure.string("Sample Description "), // description
        tx.pure.string("https://ipfs.abcdd.ai/ipfs/QmVEpxGLLfgTUjv3cwLFizaShHs48kvGthHYSJWgE5AGEk"), // url
        coin,
    ],
    

})


const result = await keyPair.signAndExecuteTransaction({
    client: suiClient,
    transaction: tx
})


// Success
// if (result.effects.status.success == "success"){

// }

await suiClient.waitForTransaction({digest: (await result).digest})


 // Failure
if (result.effects.status.error === "failure") {
    console.log("tx failed")
}

console.log("Tx executed", (await result).digest)


})()
