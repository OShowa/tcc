#include <iostream>
#include <iomanip>
#include <cstring>
#include "murmur3.h"

int main() {
    const char* input = "Hello, world!";
    const int len = std::strlen(input);
    uint32_t seed = 42;

    // Output buffer for 128-bit hash (x86 version)
    uint8_t out128[16];
    MurmurHash3_x86_128(input, len, seed, out128);

    std::cout << "MurmurHash3_x86_128(\"" << input << "\") = 0x";
    for (int i = 0; i < 16; ++i) {
        std::cout << std::hex << std::setw(2) << std::setfill('0') << (int)out128[i];
    }
    std::cout << std::dec << std::endl;

    return 0;
}