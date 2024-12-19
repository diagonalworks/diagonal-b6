import { Select } from "@/components/system/Select";

export const SelectAdapter = ({
	options,
	value,
	onValueChange,
}: {
	options: { value: string; label: string }[];
	value: string;
	onValueChange: (v: string) => void;
}) => {
	const label = (value: string) => {
		return options.find((option) => option.value === value)?.label ?? "";
	};

	return (
		<Select value={value} onValueChange={onValueChange}>
			<Select.Button>{value && label(value)}</Select.Button>
			<Select.Options>
				{options.map((option, i) => (
					<div key={i}>
						{option.value && (
							<Select.Option value={option.value}>{option.label}</Select.Option>
						)}
					</div>
				))}
			</Select.Options>
		</Select>
	);
};
