import React from 'react';

/**
 * Capitalizes the first letter of each word in a string.
 */
export const toTitleCase = (str: string) => {
    return str.replace(/\w\S*/g, function (txt) {
        return txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase();
    });
};

/**
 * Returns the text content of a React node.
 */
export const getNodeText = (node: React.ReactNode): string => {
    if (node == null) return '';

    switch (typeof node) {
        case 'string':
        case 'number':
            return node.toString();

        case 'boolean':
            return '';

        case 'object': {
            if (node instanceof Array) return node.map(getNodeText).join('');

            if ('props' in node) return getNodeText(node.props.children);
            return '';
        }

        default:
            console.warn('Unresolved `node` of type:', typeof node, node);
            return '';
    }
};

/**
 * Returns the word at a given position in a string.
 */
export const getWordAt = (str: string, pos: number) => {
    const left = str.slice(0, pos).search(/\S+$/);
    const right = str.slice(pos).search(/\s/);
    if (right < 0) {
        return str.slice(left);
    }
    return str.slice(left, right + pos);
};

/**
 * Highlights the matches in a string.
 */
export function highlighted(string: string, matches: [number, number][]) {
    const substrings = [];
    let previousEnd = 0;

    for (const [start, end] of matches) {
        const prefix = string.substring(previousEnd, start);
        const match = <strong>{string.substring(start, end)}</strong>;

        substrings.push(prefix, match);
        previousEnd = end;
    }

    substrings.push(string.substring(previousEnd));

    return <span>{React.Children.toArray(substrings)}</span>;
}
