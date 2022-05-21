from typing import Optional


def int_to_bytes(number: int) -> bytes:
    return number.to_bytes(
        length=((8 + (number + (number < 0)).bit_length()) // 8) * 4,
        byteorder="little",
        signed=True,
    )


def uint_to_bytes(number: int) -> bytes:
    return number.to_bytes(
        length=((8 + (number + (number < 0)).bit_length()) // 8) * 4,
        byteorder="little",
        signed=False,
    )


def int_from_bytes(binary_data: list) -> Optional[int]:
    return int.from_bytes(bytes(binary_data), byteorder="little", signed=True)


def uint_from_bytes(binary_data: list) -> Optional[int]:
    return int.from_bytes(bytes(binary_data), byteorder="little", signed=False)
