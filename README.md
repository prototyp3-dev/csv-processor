# CSV Processor

```
Cartesi Rollups version: 0.9.x
```

The CSV processor presents an application layer protocol that allows statements to assert facts about data, avoiding sending data to the DApp.

The DApp calculates the percentage of non-empty cells of a csv data, and sends the value together with the CID of the data. Other users can dispute the data, so the claimer has to send the data to be verified. There is no reputation system, but the DApp includes a vanity score for correct claims and won disputes. The claimer can finalize the claim, so the it is considered truthful and no one can dispute it anymore. If a disputed claim is finalized, or the claimer fails to verify, the claim is also finalized but unfavorable resolution to the claimer.

To avoid using floating point calculations, the percentage is represented by a number in [0-1000000] where 1000000 represents 1005, or no empty cells.

DISCLAIMERS

There is no reputation system or penalities for misbehaving, but we recommend them for DApps that include protocols similar to the presente in this project.

This is not a final product and should not be used as one.

## Requirements


This project works with [sunodo](https://github.com/sunodo/sunodo), so run it you should first install sunodo.

```shell
npm install -g @sunodo/cli
```

## Building

Build with:

```shell
sunodo build
```

For the frontend build with:

```shell
cd frontend
yarn
yarn codegen
```

## Running

Run with:

```shell
sunodo run
```

Run the frontend with:

```shell
yarn start
```

## Interact with the Application

Interact with the application using the web frontend

The flow of the protocol is as follows:
1. Load Wasm to be able to process data in the browser.
2. Import data: import a csv file; paste data; load from an ipfs.io CID. 
3. Process data to obtain the non-empty cells value and CID
4. Copy values to the from the import section to the output section (can optionally paste data, CID, and value directly)
5. Open a claim: send CID and value to create a claim
6. Finalize a claim: if enough time has passed, any user can send the CID to finalize an open claim, or if the claim is in dispute, can finalize the claim with an unfavorable resolt to the claimer.
7. Dispute an open claim: any user (other than the claimer) can send the CID to initiate a dispute of a claim. The claimer then should send the data to verify it.
8. Verify a claim in dispute (or open): the claimer sends the CID and the data to make the Cartesi Rollup DApp process the data and verify the claimed value.
