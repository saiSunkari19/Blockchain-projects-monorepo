import {CoinBalance, getFullnodeUrl, SuiClient} from "@mysten/sui/client"
import {MIST_PER_SUI} from "@mysten/sui/utils"
import  {getFaucetHost, requestSuiFromFaucetV2} from "@mysten/sui/faucet"
import dotenv from "dotenv"
dotenv.config();



(async ()=>{

const  MY_ADDRESS= process.env.ADDRESS!

const suiClient = new SuiClient({
    url: getFullnodeUrl('testnet')
})


const balance = (balance: CoinBalance) =>{
    return Number.parseInt(balance.totalBalance) / Number(MIST_PER_SUI)
}


// Has RATE LIMIT ISSUE get from : https://faucet.sui.io/
// await requestSuiFromFaucetV2({
//     host: getFaucetHost('testnet'), 
//     recipient: MY_ADDRESS
// })

const suiAfter = await suiClient.getBalance({
    owner: MY_ADDRESS,
    coinType: "0x02::sui::SUI"
})

console.log(
	`Balance after: ${balance(
		suiAfter,
	)} SUI. Hello, SUI!`,
);

})()
