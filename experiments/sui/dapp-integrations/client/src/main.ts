import { Ed25519Keypair} from "@mysten/sui/keypairs/ed25519"
import dotenv from "dotenv"
dotenv.config();


(async()=>{
    const keyPair = Ed25519Keypair.deriveKeypair(process.env.MNEMONIC!)
    const publicKey = keyPair.getPublicKey()

    const message = new TextEncoder().encode("Sign this message to login:\nAddress: 0xac463879634ef8be8c0db3c718a2ac9d08a1db45d48ceed3021eff656d1ce8ee\nNonce: fab4396b-70b2-48c8-be9d-d8527933e48c")
    const {signature} = await keyPair.signPersonalMessage(message)

    console.log(signature)

    const isValid = await publicKey.verifyPersonalMessage(message, signature)
    console.log(isValid)

})()


