// Copyright 2025 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
import { useCallback, useLayoutEffect, useRef, useState } from 'react';

import { Typography, TypographyProps } from '@mui/material';

interface TextOverflowMultilineProps extends TypographyProps {
  lines?: number;
  onOverflowChange?: (isOverflowing: boolean) => void;
}

const TextOverflowMultiline = ({ children, className = '', lines = 2, onOverflowChange, style, ...rest }: TextOverflowMultilineProps) => {
  const ref = useRef<HTMLElement>(null);
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [isOverflowing, setIsOverflowing] = useState(false);

  const checkOverflow = useCallback(() => {
    if (!ref.current) {
      return;
    }
    const overflowing = ref.current.scrollHeight > ref.current.clientHeight;
    setIsOverflowing(prev => {
      if (overflowing !== prev) {
        onOverflowChange?.(overflowing);
      }
      return overflowing;
    });
  }, [onOverflowChange]);

  useLayoutEffect(() => {
    checkOverflow();
  }, [checkOverflow, children, lines]);

  return (
    <Typography
      ref={ref}
      className={className}
      style={{
        display: '-webkit-box',
        WebkitLineClamp: lines,
        WebkitBoxOrient: 'vertical',
        overflow: 'hidden',
        wordBreak: 'break-word',
        ...style
      }}
      {...rest}
    >
      {children}
    </Typography>
  );
};

export default TextOverflowMultiline;
