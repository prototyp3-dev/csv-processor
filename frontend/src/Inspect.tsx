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
import { ethers } from "ethers";
// import { useRollups } from "./useRollups";

import configFile from "./config.json";

const config: any = configFile;

interface InspectMessage {
    action: string,
    id?: string
}

export const Inspect: React.FC = () => {
    // const rollups = useRollups();
    const [{ connectedChain }] = useSetChain();
    const inspectCall = async (str: string, callback: (data: string) => void ) => {
        const payload = str;

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

        fetch(`${apiURL}/${payload}`)
            .then(response => response.json())
            .then(data => {
                if (data.reports.length > 0) {
                    callback(ethers.utils.toUtf8String(data.reports[0].payload));
                }
            });
    };

    const doInspectClaimList = () => {
        const msg: InspectMessage = {action:"getClaimList"};
        inspectCall(JSON.stringify(msg),setClaimList);
    }

    const doInspectClaim = () => {
        const msg: InspectMessage = {action:"showClaim",id:inspectClaim};
        inspectCall(JSON.stringify(msg),setClaim);
    }

    const doInspectUser = () => {
        const msg: InspectMessage = {action:"showUser",id:inspectUser};
        inspectCall(JSON.stringify(msg),setUser);
    }

    const [inspectUser, setInspectUser] = useState<string>("");
    const [userData, setUser] = useState<string>("");
    const [inspectClaim, setInspectClaim] = useState<string>("");
    const [claimData, setClaim] = useState<string>("");
    const [claimList, setClaimList] = useState<string>("");

    return (
        <div>
            <div>
                All claims 
                <button onClick={() => doInspectClaimList()}>
                    Get
                </button> <br /> <br />
                {claimList}
            </div>
            <br /><br />
            <div>
                Claim <input
                    type="text"
                    value={inspectClaim}
                    onChange={(e) => setInspectClaim(e.target.value)}
                />
                <button onClick={() => doInspectClaim()}>
                    Get
                </button> <br /> <br />
                {claimData}
            </div>
            <br /> <br />

            <div>
                User <input
                    type="text"
                    value={inspectUser}
                    onChange={(e) => setInspectUser(e.target.value)}
                />
                <button onClick={() => doInspectUser()}>
                    Get
                </button> <br /> <br />
                {userData}
            </div>

        </div>
    );
};
