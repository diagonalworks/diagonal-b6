import { useEffect, useState } from 'react';

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
