import { useOutlinerContext } from '@/lib/context/outliner';
import { AtomProto, ChoiceProto } from '@/types/generated/ui';
import { AtomAdapter } from './AtomAdapter';
import { ChipAdapter } from './ChipAdapter';

export const ChoiceAdapter = ({
    choice,
}: {
    choice: {
        chips: AtomProto[];
        label: ChoiceProto['label'];
    };
}) => {
    const { choiceChips, setChoiceChipValue } = useOutlinerContext();
    return (
        <>
            {choice.label && <AtomAdapter atom={choice.label} />}
            {choice.chips &&
                choice.chips.map(({ chip }, i) => {
                    if (!chip) return null;
                    const stackChip = choiceChips[chip.index ?? 0];
                    if (stackChip) {
                        return (
                            <div key={i}>
                                {chip && (
                                    <ChipAdapter
                                        chip={stackChip}
                                        onChange={(value: number) =>
                                            setChoiceChipValue(
                                                chip.index ?? 0, // same issue with the 0 index being undefined, maybe we should add zod to parse this values beforehand or fix in BE.
                                                value
                                            )
                                        }
                                    />
                                )}
                            </div>
                        );
                    }
                })}
        </>
    );
};
