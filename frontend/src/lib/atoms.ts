import { AtomProto } from '@/types/generated/ui';
import { $FixMe } from '@/utils/defs';
import { isObject } from 'lodash';

/**
 * Recursively find atoms in a line. If a type is provided, only atoms of that type will be returned.
 * This function is currently a mess because types of line elements are loosely defined.
 */
export const findAtoms = (
    line: $FixMe,
    type?: keyof AtomProto
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
            findAtoms((line as $FixMe)[key], type)
        );
    }
    return [];
};
