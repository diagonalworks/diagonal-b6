import { isObject } from "lodash";

import { AtomProto } from "@/types/generated/ui";
import { $FixMe } from "@/utils/defs";

/**
 * Recursively find atoms in a line. If a type is provided, only atoms of that type will be returned.
 * This function is currently a mess because types of line elements are loosely defined.
 * @param line The line to search for atoms in
 * @param type The type of atom to search for
 * @returns An array of atoms found in the line
 */
export const findAtoms = (
	line: $FixMe,
	type?: keyof AtomProto,
): AtomProto[] => {
	const atom = line?.atom;
	if (atom) {
		if (type) {
			return atom?.[type] ? [atom] : [];
		}
		return [atom];
	}

	if (Array.isArray(line)) {
		return line.flatMap((l) => findAtoms(l, type));
	}

	if (isObject(line)) {
		return Object.keys(line).flatMap((key) =>
			findAtoms((line as $FixMe)[key], type),
		);
	}
	return [];
};
