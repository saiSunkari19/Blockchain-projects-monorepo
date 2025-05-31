/*
/// Module: nft_simple
module nft_simple::nft_simple;
*/

// For Move coding conventions, see
// https://docs.sui.io/concepts/sui-move-concepts/conventions

module nft_simple::nft_simple;


use std::string;
use sui::url;
use sui::url::Url;
use sui::event;

public struct NFT has key, store {
    id: UID, 
    name: string::String, 
    description: string::String, 
    url: Url
}


// Event 
public struct NFTMinted has copy, drop {
    object_id: ID, 
    creator : address, 
    name: string::String, 
    description: string::String, 
    url: Url,
}


// View Functions 

public fun  url(nft: &NFT):&Url {
    &nft.url
}


// EntryPoint
#[allow(lint(self_transfer))]
public fun mint_to_sender(
    name : vector<u8>, 
    description: vector<u8>, 
    url: vector<u8>, 
    ctx: &mut TxContext,
){
    let sender = ctx.sender(); 

    let nft = NFT{
        id: object::new(ctx), 
        name: string::utf8(name), 
        description: string::utf8(description), 
        url: url::new_unsafe_from_bytes(url), 
    };


    event::emit(NFTMinted {
        object_id: object::id(&nft),
        creator: sender,
        name: nft.name,
        description: nft.description,
        url: nft.url,
    });

    transfer::public_transfer(nft, sender);


}


// Transfer NFT
public fun transfer(nft: NFT, recipient: address, _:&mut TxContext){
    transfer::public_transfer(nft, recipient);

}


// Update the description 
public fun update_description(nft: &mut NFT, new_description: vector<u8>, _: &mut TxContext){
    nft.description = string::utf8(new_description)
}


