import { useEffect, useState } from 'react';

/**
 * Hook to determine if a text element is overflowing its container.
 * @param ref - The ref of the text element
 * @param text - The text to check for overflow
 * @returns Whether the text is overflowing
 */
export default function useOverflow(
    ref: React.RefObject<HTMLSpanElement>,
    text: string
) {
    const [isTextOverflowing, setIsTextOverflowing] = useState(false);

    useEffect(() => {
        if (ref.current) {
            setIsTextOverflowing(
                ref.current.scrollWidth > ref.current.clientWidth
            );
        }
    }, [ref, text]);

    return isTextOverflowing;
}
