import { createTeleporter } from 'react-teleporter';

export const LeftComparatorTeleporter = createTeleporter();
export const RightComparatorTeleporter = createTeleporter();

export function Comparator() {
    return (
        <div>
            <div className="flex flex-row">
                <div className="flex-grow">
                    <LeftComparatorTeleporter.Target />
                </div>
                <div className="flex-grow">
                    <RightComparatorTeleporter.Target />
                </div>
            </div>
        </div>
    );
}
