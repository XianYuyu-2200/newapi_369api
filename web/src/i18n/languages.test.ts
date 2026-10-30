/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { describe, expect, test } from 'vitest'

import {
  convertDetectedLanguage,
  normalizeInterfaceLanguage,
  toIntlLocale,
} from './languages'

describe('convertDetectedLanguage', () => {
  test('folds every detected locale onto simplified Chinese', () => {
    expect(convertDetectedLanguage('zh-TW')).toBe('zhCN')
    expect(convertDetectedLanguage('zh-HK')).toBe('zhCN')
    expect(convertDetectedLanguage('zh-MO')).toBe('zhCN')
    expect(convertDetectedLanguage('zh-Hant-TW')).toBe('zhCN')
    expect(convertDetectedLanguage('zhTW')).toBe('zhCN')
    expect(convertDetectedLanguage('zh')).toBe('zhCN')
    expect(convertDetectedLanguage('zh-CN')).toBe('zhCN')
    expect(convertDetectedLanguage('zh-Hans')).toBe('zhCN')
    expect(convertDetectedLanguage('zhCN')).toBe('zhCN')
    expect(convertDetectedLanguage('en')).toBe('zhCN')
    expect(convertDetectedLanguage('fr-FR')).toBe('zhCN')
    expect(convertDetectedLanguage('ja')).toBe('zhCN')
  })
})

describe('normalizeInterfaceLanguage', () => {
  test('keeps only simplified Chinese', () => {
    expect(normalizeInterfaceLanguage()).toBe('zhCN')
    expect(normalizeInterfaceLanguage('')).toBe('zhCN')
    expect(normalizeInterfaceLanguage('en')).toBe('zhCN')
    expect(normalizeInterfaceLanguage('zhTW')).toBe('zhCN')
    expect(normalizeInterfaceLanguage('zh-TW')).toBe('zhCN')
    expect(normalizeInterfaceLanguage('zhCN')).toBe('zhCN')
  })
})

describe('toIntlLocale', () => {
  test('maps the interface code to a BCP-47 tag', () => {
    expect(toIntlLocale('zhCN')).toBe('zh-CN')
  })
})
