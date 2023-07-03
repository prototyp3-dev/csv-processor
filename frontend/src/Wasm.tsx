// Copyright 2022 Cartesi Pte. Ltd.

// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the license at http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

import React, { useState } from "react";
import { useSetChain } from "@web3-onboard/react";
import configFile from "./config.json";

require('./wasm_exec');

const config: any = configFile;

export const Wasm: React.FC = () => {
    // const rollups = useRollups();
    const hexToBytes = (hex:string) => {
        let bytes = [];
        for (let c = 0; c < hex.length; c += 2) {
            bytes.push(parseInt(hex.substring(c, c+2), 16));
        }
        return new Uint8Array(bytes);
    }

    const go = new (window as any).Go();
    
    const [{ connectedChain }] = useSetChain();
    const inspectCall = async () => {
        const payload = '{"action":"wasm"}';

        if (!connectedChain){
            return;
        }
        
        let apiURL= ""

        if(config[connectedChain.id]?.inspectAPIURL) {
            apiURL = `${config[connectedChain.id].inspectAPIURL}/inspect`;
        } else {
            console.error(`No inspect interface defined for chain ${connectedChain.id}`);
            return;
        }
        setWasmMessage("Loading wasm...");

        fetch(`${apiURL}/${payload}`)
            .then(response => response.json())
            .then(data => {
                if (data.reports.length > 0) {
                    for (let i = 0; i < data.reports.length; i += 1) {
                        const wasmBytes = hexToBytes(data.reports[i].payload.substring(2));
                        
                        //remove the message: syscall/js.finalizeRef not implemented
                        go.importObject.env["syscall/js.finalizeRef"] = () => {}

                        WebAssembly.instantiate(wasmBytes, go.importObject).then( (obj:any) => {
                            const wasm = obj.instance;
                            go.run(wasm);
                            setWasmMessage("Wasm loaded!");
                        })
                    }
                }
        
            }).catch(() => {
                setWasmMessage("Something went wrong...");
            });
    };
    const [wasmMessage, setWasmMessage] = useState<string>("");

    return (
        <div>
            <div>
                <button onClick={() => inspectCall()}>
                    Load Wasm
                </button>
                <span> {wasmMessage}</span>
            </div>
        </div>
    );
};
