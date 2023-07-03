// Copyright 2022 Cartesi Pte. Ltd.

// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the license at http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

import React, { useRef,useState } from "react";
import { ethers } from "ethers";
import { useRollups } from "./useRollups";

interface IInputPropos {
    dappAddress: string 
}

interface ClaimMessage {
    action: string,
    id: string,
    value?: number,
    data?: string
}

export const Input: React.FC<IInputPropos> = (propos) => {
    const rollups = useRollups(propos.dappAddress);
    const fileRef = useRef<HTMLInputElement | null>(null);

    const addInput = async (str: string) => {
        if (rollups) {
            rollups.inputContract.addInput(rollups.dappContract.address, ethers.utils.toUtf8Bytes(str))
                .catch((e) => {
                    console.log(e)
            });
        }
    };

    const copyProcessedValues = () => {
        setCidToSend(cid);
        setValuePermillionToSend(valuePermillion);
        setCsvDataToSend(csvData);
    }

    const sendClaim = () => {
        const claim: ClaimMessage = {action:"claim",id:cidToSend,value:valuePermillionToSend};
        addInput(JSON.stringify(claim));
    }

    const sendFinalize = () => {
        const claim: ClaimMessage = {action:"finalize",id:cidToSend};
        addInput(JSON.stringify(claim));
    }

    const sendDispute = () => {
        const claim: ClaimMessage = {action:"dispute",id:cidToSend};
        addInput(JSON.stringify(claim));
    }

    const sendValidate = () => {
        if (csvDataToSend.length > maxSizeToSend) {
            const chunks = (window as any).prepareData(csvDataToSend,maxSizeToSend);
            for (let c = 0; c < chunks.length; c += 1) {
                const chunkToSend = chunks[c];
                const claim: ClaimMessage = {action:"validateChunk",id:cidToSend,data:chunkToSend};
                addInput(JSON.stringify(claim));
            }
        } else {
            const claim: ClaimMessage = {action:"validate",id:cidToSend,data:csvDataToSend};
            addInput(JSON.stringify(claim));
        }
    }

    const processCsv = () => {
        setCid("");
        setValuePermillion(0);
        try {
            setCid((window as any).getDataCid(csvData));
            setValuePermillion((window as any).emptyCellValue(csvData));
        } catch(e) {
            console.log(e)
        }
    }
    const handleOnChange = (e: any) => {
        const reader = new FileReader();
        reader.onload = async (readerEvent) => {
            const text = readerEvent.target?.result;
            if (text) {
                setCsvData(String(text));
                e.target.value = null
            }
        };
        reader.readAsText(e.target.files[0])
    }

    const readFile = () => {
        fileRef.current?.click();
    };

    const loadFromIpfs = () => {
        setIpfsLoading(true);
        fetch(`https://ipfs.io/ipfs/${cidIpfs}`)
            .then(response => response.text())
            .then(data => {
                setIpfsLoading(false);
                setCsvData(String(data));
            });
    };

    const [ipfsLoading, setIpfsLoading] = useState<boolean>(false);
    const [cidIpfs, setCidIpfs] = useState<string>("");
    const [csvData, setCsvData] = useState<string>("");
    const [csvDataToSend, setCsvDataToSend] = useState<string>("");
    const [cid, setCid] = useState<string>("");
    const [valuePermillion, setValuePermillion] = useState<number>(0);
    const [cidToSend, setCidToSend] = useState<string>("");
    const [valuePermillionToSend, setValuePermillionToSend] = useState<number>(0);
    const [maxSizeToSend, setmaxSizeToSend] = useState<number>(409600);

    return (
        <div>
            <h3>Data</h3>
            <div>
                CSV data 
                <br />
                <br />
                <input type="file" accept=".csv" ref={fileRef} onChange={(e) => handleOnChange(e)} style={{ display: 'none' }}/>
                <button onClick={() => readFile()}>
                    Read File
                </button><span>&nbsp;&nbsp;&nbsp;&nbsp;</span>
                <button onClick={() => loadFromIpfs()} disabled={ipfsLoading}>
                    Load From ipfs.io
                </button>
                <input type="text" value={cidIpfs} onChange={(e) => setCidIpfs(e.target.value)} />
                <br />
                <textarea
                    value={csvData}
                    onChange={(e) => setCsvData(e.target.value)}
                />
                <br />
                <button onClick={() => processCsv()} disabled={!((window as any).getDataCid && (window as any).emptyCellValue && csvData)}>
                    Process
                </button>
                {!((window as any).getDataCid && (window as any).emptyCellValue) && <span> Load wasm to process CSV</span> }
                
            </div>
            <br />
            <div>
                CID: <input
                    disabled
                    type="text"
                    value={cid}
                    onChange={(e) => setCid(e.target.value)}
                />
                <br/>
                VALUE per 1000000: <input
                    disabled
                    type="number"
                    min="0"
                    max="1000000"
                    value={valuePermillion}
                    onChange={(e) => setValuePermillion(Number(e.target.value))}
                /> 
            </div>

            <h3>Interact with DApp</h3>
            <button onClick={() => copyProcessedValues()}>
                Copy Processed Values
            </button>
            <br />
            <br />
            <div>
                CID To Send: <input
                    type="text"
                    value={cidToSend}
                    onChange={(e) => setCidToSend(e.target.value)}
                />
                <br/>
                <br/>
                VALUE Per 1000000 to send <input
                    type="number"
                    min="0"
                    max="1000000"
                    value={valuePermillionToSend}
                    onChange={(e) => setValuePermillionToSend(Number(e.target.value))}
                /> 
                <br/>
                <br/>
                CSV data to send 
                <br />
                <textarea
                    value={csvDataToSend}
                    onChange={(e) => setCsvDataToSend(e.target.value)}
                />
            </div>
            <h4>Send [values] to DApp</h4>
            <div>
                <button onClick={() => sendClaim()} disabled={!rollups}>
                    Claim
                </button> (Claim CID and value per 1000000 - sends CID and VALUE)
                <br /><br />
                <button onClick={() => sendFinalize()} disabled={!rollups}>
                    Finalize
                </button> (Finalize either an undisputed claim or an unvalidated disputed claim - sends CID)
                <br /><br />
                <button onClick={() => sendDispute()} disabled={!rollups}>
                    Dispute
                </button> (Dispute an open claim - sends CID)
                <br /><br />
                <button onClick={() => sendValidate()} disabled={!rollups}>
                    Validate
                </button> (Validate claimed CID with whole data - sends CID and CSV)
                <span>   -  Max chunk size: </span><input
                    type="number"
                    min="0"
                    value={maxSizeToSend}
                    onChange={(e) => setmaxSizeToSend(Number(e.target.value))}
                /> 
                <br /><br />
            </div>
        </div>
    );
};
