import { ChipProto, UIResponseProto } from './generated/ui';

export type StackResponse = {
    proto: UIResponseProto;
};

// enriched Chip type that includes the selected value of the chip
export type Chip = { atom: ChipProto; value: number };
