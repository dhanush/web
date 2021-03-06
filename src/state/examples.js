import * as history from '../services/history';
import * as code from './code';
import * as languages from './languages';

import { java_example_1 } from '../examples/hello.java.js';
import { java_example_2 } from '../examples/swap.java.js';
import { python_example_1 } from '../examples/fizzbuzz.py.js';
import { python_example_2 } from '../examples/classdef.py.js';
import { go_example_1 } from '../examples/golang.go.js';
import { php_example_1 } from '../examples/phphtml.php.js';
import { php_example_2 } from '../examples/phpv7.php.js';
import { javascript_example_1 } from '../examples/javascript.js.js';
import { ruby_example_1 } from '../examples/ruby.rb.js';
import { bash_example_1 } from '../examples/bash.sh.js';
import { csharp_example_1 } from '../examples/csharp.cs.js';

export const DEFAULT_EXAMPLE = 'java_example_1';
const LANG_JAVA = 'java';
const LANG_PYTHON = 'python';
const LANG_GO = 'go';
const LANG_PHP = 'php';
const LANG_JS = 'javascript';
const LANG_RUBY = 'ruby';
const LANG_BASH = 'bash';
const LANG_CSHARP = 'csharp';

const examples = {
  java_example_1: {
    name: 'hello.java',
    language: LANG_JAVA,
    code: java_example_1,
  },
  java_example_2: {
    name: 'swap.java',
    language: LANG_JAVA,
    code: java_example_2,
  },
  python_example_1: {
    name: 'fizzbuzz.py',
    language: LANG_PYTHON,
    code: python_example_1,
  },
  python_example_2: {
    name: 'classdef.py',
    language: LANG_PYTHON,
    code: python_example_2,
  },
  go_example_1: {
    name: 'golang.go',
    language: LANG_GO,
    code: go_example_1,
  },
  php_example_1: {
    name: 'phphtml.php',
    language: LANG_PHP,
    code: php_example_1,
  },
  php_example_2: {
    name: 'phpv7.php',
    language: LANG_PHP,
    code: php_example_2,
  },
  javascript_example_1: {
    name: 'javascript.js',
    language: LANG_JS,
    code: javascript_example_1,
  },
  ruby_example_1: {
    name: 'ruby.rb',
    language: LANG_RUBY,
    code: ruby_example_1,
  },
  bash_example_1: {
    name: 'bash.sh',
    language: LANG_BASH,
    code: bash_example_1,
    driver: LANG_BASH,
  },
  csharp_example_1: {
    name: 'csharp.cs',
    language: LANG_CSHARP,
    code: csharp_example_1,
  },
};

export const initialState = {
  list: {},
  selected: '',
};

export const SET = 'bblfsh/examples/SET';
export const RESET = 'bblfsh/examples/RESET';
export const SET_LIST_BY_LANGS = 'bblfsh/examples/SET_LIST_BY_LANGS';

export const reducer = (state = initialState, action) => {
  switch (action.type) {
    case SET:
      return {
        ...state,
        selected: action.selected,
      };
    case RESET:
      return {
        ...state,
        selected: '',
      };
    case SET_LIST_BY_LANGS:
      const { languages } = action;
      const list = Object.keys(examples).reduce((acc, key) => {
        const example = examples[key];
        if (languages.includes(example.language)) {
          acc[key] = example;
        }
        return acc;
      }, {});
      return {
        ...state,
        list,
      };
    default:
      return state;
  }
};

export const select = key => dispatch => {
  if (!key) {
    return;
  }

  const example = examples[key];

  dispatch(code.set(example.name, example.code));

  const forcedDriver = example.driver;
  dispatch(languages.select(forcedDriver));
  if (!forcedDriver) {
    dispatch(languages.set(example.language));
  }

  // url side effect
  history.setExample();

  return dispatch({
    type: SET,
    selected: key,
  });
};

export const reset = () => ({ type: RESET });

export const setListByLangs = languages => ({
  type: SET_LIST_BY_LANGS,
  languages,
});
