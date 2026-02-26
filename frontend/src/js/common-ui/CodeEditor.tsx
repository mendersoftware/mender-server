// Copyright 2026 Northern.tech AS
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
import { useCallback, useEffect, useRef } from 'react';
import { useSelector } from 'react-redux';

import { alpha, rgbToHex, useTheme } from '@mui/material';
import { makeStyles } from 'tss-react/mui';

import MonacoEditor, { EditorProps, Monaco, loader } from '@monaco-editor/react';
import { getIsDarkMode } from '@northern.tech/store/selectors';
import { editor } from 'monaco-editor';

import Loader from './Loader';

loader.config({ paths: { vs: '/ui/vs' } });

const useStyles = makeStyles()(theme => {
  const isDarkMode = theme.palette.mode == 'dark';
  return {
    wrapper: {
      border: `1px solid ${theme.palette.action.active}`,
      borderRadius: theme.spacing(0.5),
      '> section': { minHeight: '20lh' },
      '.monaco-editor, .monaco-editor-background, .monaco-editor .margin': {
        backgroundColor: 'transparent'
      },
      '.monaco-editor .view-overlays .current-line-exact': {
        border: isDarkMode ? `1px solid ${theme.palette.grey[800]}` : 'inherit'
      },
      '&.readOnly': {
        borderColor: isDarkMode ? alpha(theme.palette.action.active, 0.56) : theme.palette.divider,
        '.monaco-editor-background': {
          backgroundColor: isDarkMode ? alpha(theme.palette.info.light, theme.palette.action.selectedOpacity) : theme.palette.action.hover
        }
      }
    }
  };
});

export const defaultEditorOptions: EditorProps['options'] = {
  autoClosingOvertype: 'auto',
  codeLens: false,
  contextmenu: false,
  enableSplitViewResizing: false,
  fontFamily: 'Red Hat Mono',
  fontSize: 14,
  formatOnPaste: true,
  lightbulb: { enabled: 'off' },
  minimap: { enabled: false },
  overviewRulerBorder: false,
  padding: { top: 16, bottom: 16 },
  quickSuggestions: false,
  renderOverviewRuler: false,
  scrollBeyondLastLine: false,
  wordWrap: 'on'
};

const makeColorMonacoSafe = (color: string) => {
  const hexColor = rgbToHex(color);
  return hexColor.length < 6 ? `${hexColor}${hexColor.substring(1)}` : hexColor;
};

export const useEditorTheme = (isReadOnly: boolean) => {
  const monacoRef = useRef<Monaco | null>(null);
  const isDarkMode = useSelector(getIsDarkMode);
  const muiTheme = useTheme();

  const editorThemeName = `mender-${muiTheme.palette.mode}${isReadOnly ? '-ro' : ''}`;

  const defineEditorTheme = useCallback(
    (monaco: Monaco) => {
      monacoRef.current = monaco;
      monaco.editor.defineTheme(editorThemeName, {
        base: isDarkMode ? 'vs-dark' : 'vs',
        inherit: true,
        rules: [],
        colors: { 'editor.foreground': makeColorMonacoSafe(muiTheme.palette.text.primary) }
      });
    },
    [isDarkMode, muiTheme, editorThemeName]
  );

  useEffect(() => {
    if (!monacoRef.current) {
      return;
    }
    defineEditorTheme(monacoRef.current);
    monacoRef.current.editor.setTheme(editorThemeName);
  }, [defineEditorTheme, editorThemeName]);

  return { editorThemeName, defineEditorTheme };
};

interface CodeEditorProps {
  className?: string;
  language: string;
  onChange?: (value: string | undefined, ev: editor.IModelContentChangedEvent) => void;
  onMount?: (editor: editor.IStandaloneCodeEditor, monaco: Monaco) => void;
  options?: EditorProps['options'];
  readOnly?: boolean;
  value: string;
}

export const CodeEditor = ({ className = '', language, onChange, onMount, options = {}, readOnly = false, value }: CodeEditorProps) => {
  const { classes } = useStyles();
  const { editorThemeName, defineEditorTheme } = useEditorTheme(readOnly);

  const handleEditorDidMount = (editor: editor.IStandaloneCodeEditor, monaco: Monaco) => {
    defineEditorTheme(monaco);
    onMount?.(editor, monaco);
  };

  return (
    <div className={`margin-top-small margin-bottom-small ${classes.wrapper} ${readOnly ? 'readOnly' : ''} ${className}`}>
      <MonacoEditor
        beforeMount={defineEditorTheme}
        defaultLanguage={language}
        language={language}
        loading={<Loader show />}
        onChange={onChange}
        onMount={handleEditorDidMount}
        options={{
          ...defaultEditorOptions,
          ...options,
          readOnly
        }}
        theme={editorThemeName}
        value={value}
        wrapperProps={{ 'data-testid': 'monaco-editor' }}
      />
    </div>
  );
};

export default CodeEditor;
