# AMD States Linux
Dynamically change AMD Ryzen Processor C-State and P-States

## Overview
This tool provides a command-line interface (CLI) for manipulating Model-Specific Registers (MSRs) on AMD Ryzen processors. It allows users to adjust performance states (P-States) and power-saving states (C-States) directly through MSR reads/writes, offering a means to fine-tune processor performance and power usage characteristics.

## Features

- P-State Management: Modify the processor's performance states to optimize power consumption and performance based on your needs.
- C-State Management: Control idle power-saving states to enhance battery life or reduce energy consumption.
- Dynamic Listing: View the current P-State and C-State configurations directly from your terminal.
- Compatibility: Designed for AMD Ryzen processors, but may work with other processors that utilize similar MSR addresses for P-States and C-States.

## Prerequisites

- AMD Ryzen processor (other processors may be compatible, but are not explicitly supported)
- Linux operating system with MSR module loaded (modprobe msr)
- Root access or sufficient privileges to read/write MSRs
- Go programming environment for building the tool

## Building from Source
Clone the repository and build the tool using Go:

```bash
git clone https://github.com/ZanMax/amd-states-linux.git
cd amd-states-linux
go build -o msr-tool main.go
```

## Usage
Run the tool with root privileges to ensure access to MSRs. Use the -h flag to display help and available commands:

```bash
sudo ./msr-tool -h
```

## Example Commands
List all P-States:
```bash
sudo ./msr-tool -list
```

Enable a specific P-State:
```bash
sudo ./msr-tool -pstate 0 --enable
```

Disable C6 state:
```bash
sudo ./msr-tool --c6-disable
```

## Safety and Disclaimer
Modifying MSRs can affect system stability and performance. Use this tool at your own risk. Always ensure you have a way to recover your system if something goes wrong. The developers of this tool are not responsible for any damage or data loss that may occur from its use.

