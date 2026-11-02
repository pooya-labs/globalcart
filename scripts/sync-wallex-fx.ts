import { randomUUID } from "node:crypto";
import { readFile, rename, unlink, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { normalizeWallexMinuteHistory } from "../lib/fx/wallex";
const endpoint = "https://api.wallex.ir/v1/udf/history";
const outputPath = join(process.cwd(), "data", "fx-rate.json");
async function main(): Promise<void> {
 const nowSeconds=Math.floor(Date.now()/1000);
 const url=new URL(endpoint);
 url.searchParams.set("symbol","USDTTMN"); url.searchParams.set("resolution","1");
 url.searchParams.set("from",String(nowSeconds-3600)); url.searchParams.set("to",String(nowSeconds));
 const response=await fetch(url,{headers:{accept:"application/json","user-agent":"Mozilla/5.0 (compatible; GlobalCart-FX-Snapshot/1.0)"},signal:AbortSignal.timeout(15_000),cache:"no-store"});
 if(!response.ok)throw new Error("Wallex UDF request failed with HTTP "+response.status);
 const quote=normalizeWallexMinuteHistory(await response.json());
 const serialized=JSON.stringify(quote,null,2)+"\n";
 let previous="";
 try{previous=await readFile(outputPath,"utf8");}catch(error){
  if(!(error&&typeof error==="object"&&"code"in error&&error.code==="ENOENT"))throw error;
 }
 if(previous===serialized){process.stdout.write("Wallex USDTTMN FX snapshot is unchanged; no commit is needed.\n");return;}
 const temporaryPath=outputPath+"."+randomUUID()+".tmp";
 try{await writeFile(temporaryPath,serialized,{encoding:"utf8",flag:"wx"});await rename(temporaryPath,outputPath);}
 finally{await unlink(temporaryPath).catch(()=>undefined);}
 process.stdout.write("Updated Wallex USDTTMN 1-minute candle close to "+Math.round(quote.tomanPerUsdt??0).toLocaleString("en-US")+" Toman per USDT; candle started at "+quote.candleStartedAt+".\n");
}
main().catch(error=>{process.stderr.write((error instanceof Error?error.message:String(error))+"\n");process.exitCode=1;});
