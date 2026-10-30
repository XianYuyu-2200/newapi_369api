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

export type DocsAppPath =
  | '/'
  | '/codex'
  | '/dashboard'
  | '/docs'
  | '/group-status'
  | '/keys'
  | '/pricing'
  | '/sign-in'
  | '/sign-up'

export type DocsTarget =
  | { type: 'app'; to: DocsAppPath }
  | { type: 'article'; section: string; article: string }
  | { type: 'external'; href: string }
  | { type: 'section'; section: string }

export type DocsBlock =
  | { type: 'code'; text: string }
  | { type: 'h2'; id: string; text: string }
  | { type: 'image'; alt: string; caption: string; src: string }
  | { type: 'note'; text: string }
  | { type: 'ol'; items: string[] }
  | { type: 'p'; text: string }
  | { type: 'ul'; items: string[] }

export type DocsAction = {
  label: string
  target: DocsTarget
}

export type DocsArticle = {
  action?: DocsAction
  blocks: DocsBlock[]
  slug: string
  summary: string
  title: string
}

export type DocsGroup = {
  articles: DocsArticle[]
  id: string
  title: string
}

export type DocsIcon = 'book' | 'code' | 'help' | 'terminal'

export type DocsSection = {
  action?: DocsAction
  description: string
  groups: DocsGroup[]
  heading: string
  icon: DocsIcon
  id: string
  title: string
}

export const docsHome = {
  description:
    '从注册账号、创建密钥到配置客户端，按当前要做的事选择入口。',
  eyebrow: '369api API Documentation',
  headline: '只需站点地址和 API 密钥，即可完成模型配置',
  title: '369api 文档',
}

const article = (
  slug: string,
  title: string,
  summary: string,
  blocks: DocsBlock[],
  action?: DocsAction
): DocsArticle => ({ action, blocks, slug, summary, title })

const shot = (file: string, alt: string, caption: string): DocsBlock => ({
  alt,
  caption,
  src: `/guide/${file}`,
  type: 'image',
})

export const docsSections: DocsSection[] = [
  {
    id: 'start',
    title: '使用教程',
    heading: '开始使用',
    description: '先安装 Codex，再注册账号并创建 API 密钥。',
    icon: 'book',
    action: {
      label: '从第一步开始',
      target: { type: 'article', section: 'start', article: 'codex-install' },
    },
    groups: [
      {
        id: 'codex-install',
        title: 'Codex 安装教程',
        articles: [
          article(
            'codex-install',
            '安装 Codex',
            '从主站的 Codex 安装包页面安装客户端，再继续注册账号和创建密钥。',
            [
              {
                type: 'note',
                text: '安装文件以「Codex安装包」页面实际列出的为准。这里不写下载地址或版本号。',
              },
              { type: 'h2', id: 'get-package', text: '获取安装包' },
              {
                type: 'ol',
                items: [
                  '打开主站导航里的「Codex安装包」。',
                  '按自己的系统选择页面上列出的安装文件。',
                  '页面还没有对应文件时，先不要另找下载地址。',
                ],
              },
              shot(
                'codex-nav.png',
                '顶部导航里圈出的 Codex安装包',
                '红圈是顶部导航里的「Codex安装包」。'
              ),
              shot(
                'codex-page.png',
                'Codex安装包页面，红圈标出页面标题',
                '打开后目前只有页面标题，没有列出安装文件。文件以页面实际显示的为准。'
              ),
              { type: 'h2', id: 'install-client', text: '安装并打开' },
              {
                type: 'ol',
                items: [
                  '按系统提示完成安装。',
                  '安装完成后打开 Codex，确认程序能启动。',
                  '不要在客户端里填写上游渠道密钥。',
                ],
              },
              { type: 'h2', id: 'after-install', text: '安装之后' },
              {
                type: 'ul',
                items: [
                  '还没有账号时，继续看「注册账号」。',
                  '然后在控制台创建 API 密钥，并选对分组。',
                  '把站点接口地址、密钥和模型名填进 Codex。填写位置以当前客户端版本为准，也可以先看「客户端配置」里的 Codex 说明。',
                ],
              },
            ],
            { label: '打开安装包页面', target: { type: 'app', to: '/codex' } }
          ),
        ],
      },
      {
        id: 'account',
        title: '账号与密钥',
        articles: [
          article(
            'register',
            '注册账号',
            '创建 369api 账号后，才能进入控制台管理密钥。',
            [
              {
                type: 'note',
                text: '站点实际开放的注册和登录方式，以注册页显示为准。',
              },
              { type: 'h2', id: 'open-sign-up', text: '打开注册页' },
              {
                type: 'p',
                text: '在主站点击「注册」，按页面要求填写账号信息并完成注册。',
              },
              shot(
                'register-nav.png',
                '顶部导航里圈出的注册按钮',
                '红圈是首页右上角的「注册」。'
              ),
              shot(
                'signup-form.png',
                '注册表单，红圈标出用户名、密码、确认密码和创建账户',
                '红圈是用户名、密码、确认密码和「创建账户」。注册页若还有其他字段，以页面实际显示为准。'
              ),
              { type: 'h2', id: 'sign-in', text: '登录控制台' },
              {
                type: 'ol',
                items: [
                  '使用刚注册的账号登录。',
                  '进入控制台。',
                  '之后就可以创建 API 密钥、查看模型广场和分组状态。',
                ],
              },
              shot(
                'signin-form.png',
                '登录表单，红圈标出账号、密码和登录按钮',
                '红圈是登录用的账号、密码和「登录」。'
              ),
            ],
            { label: '去注册', target: { type: 'app', to: '/sign-up' } }
          ),
          article(
            'api-key',
            '创建 API 密钥',
            '客户端和自己的程序都使用这里创建的密钥调用接口。',
            [
              {
                type: 'note',
                text: '完整密钥通常只在创建时显示一次。关掉页面前请自行保存，不要发给别人。',
              },
              { type: 'h2', id: 'create-key', text: '创建步骤' },
              {
                type: 'ol',
                items: [
                  '登录后打开控制台里的「API 密钥」。',
                  '新建一把密钥，并按页面要求选择分组等选项。',
                  '复制完整密钥，填到客户端或请求头里。',
                ],
              },
              shot(
                'keys-create.png',
                '控制台 API 密钥页，红圈标出侧栏入口和创建按钮',
                '红圈分别是侧栏「API 密钥」和右上角「创建 API 密钥」。图中没有真实密钥。'
              ),
              shot(
                'keys-drawer.png',
                '创建 API 密钥抽屉，红圈标出名称和分组',
                '红圈是「名称」和「分组」。分组决定这把密钥可以调用哪些模型。图中没有提交，也没有密钥内容。'
              ),
              { type: 'h2', id: 'keep-key', text: '保管密钥' },
              {
                type: 'ul',
                items: [
                  '不要把密钥提交到公开仓库。',
                  '不要把密钥写进会发给用户的前端页面。',
                  '密钥泄露后，在同一页面停用或删除它，再创建新的。',
                ],
              },
            ],
            { label: '打开 API 密钥', target: { type: 'app', to: '/keys' } }
          ),
          article(
            'groups',
            '分组与模型',
            '密钥选择的分组，决定这把密钥可以调用哪些模型。',
            [
              {
                type: 'p',
                text: '不同分组可用的模型和价格可能不同。价格以模型广场当前显示为准，文档里不另写一份，避免和页面不一致。',
              },
              { type: 'h2', id: 'choose-group', text: '先看哪里' },
              {
                type: 'ul',
                items: [
                  '模型广场：查看模型和价格。',
                  '创建密钥时：选择这把密钥使用的分组。',
                  '分组状态：查看各分组当前是否可用。',
                ],
              },
              shot(
                'pricing-nav.png',
                '顶部导航里圈出的模型广场',
                '红圈是顶部导航里的「模型广场」。'
              ),
              shot(
                'pricing-group.png',
                '模型广场筛选里圈出的分组',
                '红圈是模型广场左侧的分组筛选。价格、倍率和有哪些模型，以这一页当前显示为准，不要把图里的数字当成固定报价。'
              ),
              shot(
                'group-status.png',
                '控制台分组状态页，红圈标出侧栏入口',
                '红圈是控制台侧栏里的「分组状态」。可用情况来自近期请求，不是固定公告。'
              ),
              {
                type: 'note',
                text: '调用失败时，除了检查地址和密钥，也要确认这把密钥的分组里包含所请求的模型。',
              },
            ],
            { label: '打开模型广场', target: { type: 'app', to: '/pricing' } }
          ),
        ],
      },
      {
        id: 'first-use',
        title: '第一次调用',
        articles: [
          article(
            'first-call',
            '发起第一次调用',
            '在要用的 API 密钥这一行点 CC Switch，把配置导入客户端，不用手填接口地址。',
            [
              { type: 'h2', id: 'open-key', text: '点这把密钥的 CC Switch' },
              {
                type: 'p',
                text: '先安装 CC Switch，并在控制台创建好 API 密钥。导入时打开「API 密钥」，找到要用的那一把，点这一行编辑按钮旁边的 CC Switch。它不在右侧省略号菜单里。',
              },
              shot(
                'ccswitch-key.png',
                'API 密钥列表中圈出的 CC Switch 按钮',
                '红圈是这把密钥旁边的 CC Switch，图标是左右箭头，紧挨在编辑按钮右侧。图里的密钥字符串已隐藏。'
              ),
              { type: 'h2', id: 'import', text: '在对话框里导入' },
              {
                type: 'ol',
                items: [
                  '选择应用。配置 Codex 就选 Codex，也可以选 Claude 或 Gemini。',
                  '填写在 CC Switch 里显示的名称。',
                  '选择主模型。选 Claude 时还可以填 Haiku、Sonnet、Opus 模型，这三项可以不填。',
                  '点「打开 CC Switch」。已安装的客户端会打开，并带上这把密钥和站点接口地址。',
                ],
              },
              shot(
                'ccswitch-import.png',
                '填入 CC Switch 对话框，红圈标出应用、主模型和打开按钮',
                '红圈从上到下是应用、主模型和「打开 CC Switch」。主模型选这把密钥当前分组里可用的模型，有哪些以模型广场显示为准。'
              ),
              {
                type: 'note',
                text: '这一步只把配置交给本机的 CC Switch，不会修改密钥。还没安装客户端时，用顶部导航的 CCSwitch 打开官网。不使用 CC Switch、自己写程序时，继续看「开发接入」。',
              },
            ],
            { label: '打开 API 密钥', target: { type: 'app', to: '/keys' } }
          ),
        ],
      },
    ],
  },
  {
    id: 'clients',
    title: '客户端配置',
    heading: '客户端配置',
    description: '把站点地址和 API 密钥填进客户端。具体字段可以继续补到对应文章里。',
    icon: 'terminal',
    action: {
      label: '配置 CCSwitch',
      target: { type: 'article', section: 'clients', article: 'ccswitch' },
    },
    groups: [
      {
        id: 'clients',
        title: '常用客户端',
        articles: [
          article(
            'ccswitch',
            'CCSwitch',
            '安装 CC Switch 后，在 API 密钥这一行导入配置，不用手填接口地址。',
            [
              {
                type: 'p',
                text: '顶部导航的 CCSwitch 会打开官网，用来安装客户端。接入 369api 不要在官网里手填地址，回到控制台，在对应 API 密钥这一行点 CC Switch。步骤见「发起第一次调用」。',
              },
              shot(
                'ccswitch-nav.png',
                '顶部导航里圈出的 CCSwitch',
                '红圈打开 CCSwitch 官网，用来安装客户端。导入配置在 API 密钥这一行，不在这个官网页面里。'
              ),
              { type: 'h2', id: 'need', text: '导入时确认' },
              {
                type: 'ul',
                items: [
                  '已经安装 CC Switch。',
                  '控制台里有一把要用的 API 密钥。',
                  '对话框里的主模型属于这把密钥的分组，以模型广场当前显示为准。',
                ],
              },
              {
                type: 'note',
                text: '点「打开 CC Switch」后，地址和密钥由控制台交给本机客户端，不用自己粘贴。',
              },
            ],
            {
              label: '打开 CCSwitch 官网',
              target: { type: 'external', href: 'https://ccswitch.io/zh/' },
            }
          ),
          article(
            'codex',
            'Codex',
            '安装 Codex 后，用 369api 的接口地址和密钥完成配置。',
            [
              {
                type: 'p',
                text: '安装步骤见使用教程里的「Codex 安装教程」。安装包从主站的「Codex安装包」进入。配置时准备下面三项。',
              },
              shot(
                'codex-nav.png',
                '顶部导航里圈出的 Codex安装包',
                '红圈是安装包入口。客户端里具体填在哪一栏，以当前版本为准。'
              ),
              { type: 'h2', id: 'codex-need', text: '配置时需要' },
              {
                type: 'ul',
                items: [
                  'Codex 安装包或官方安装方式。',
                  '369api 的接口地址。',
                  '对应分组的 API 密钥和模型名。',
                ],
              },
            ],
            { label: '打开安装包页面', target: { type: 'app', to: '/codex' } }
          ),
          article(
            'compatible',
            '其他 OpenAI 兼容客户端',
            '支持自定义接口地址和 Bearer 密钥的客户端，可以按同一套信息接入。',
            [
              { type: 'h2', id: 'fields', text: '通常要填的三项' },
              {
                type: 'ul',
                items: [
                  'Base URL：站点地址加上 /v1。',
                  'API Key：控制台创建的密钥。',
                  'Model：模型广场中、当前分组可用的模型名。',
                ],
              },
              shot(
                'endpoint.png',
                '首页上圈出的 OpenAI 兼容地址',
                '红圈是首页给出的地址。Base URL 通常就是它；如果客户端会自己追加 /v1，先核对后再填。'
              ),
              {
                type: 'note',
                text: '有的客户端会自己在 Base URL 后面追加 /v1 或 /chat/completions。如果请求路径重复或缺少，先核对客户端的拼接规则。',
              },
            ],
            {
              label: '查看接口地址',
              target: {
                type: 'article',
                section: 'developers',
                article: 'endpoint',
              },
            }
          ),
        ],
      },
    ],
  },
  {
    id: 'developers',
    title: '开发接入',
    heading: '开发接入',
    description: '在自己的程序里使用站点地址、密钥和模型名调用接口。',
    icon: 'code',
    action: {
      label: '查看接口地址',
      target: { type: 'article', section: 'developers', article: 'endpoint' },
    },
    groups: [
      {
        id: 'api',
        title: '接口说明',
        articles: [
          article(
            'endpoint',
            '接口地址',
            '聊天补全请求发到站点自己的 /v1/chat/completions。',
            [
              {
                type: 'p',
                text: '把下面的主机名换成你正在使用的站点地址。本地预览和生产域名不要混用。',
              },
              {
                type: 'code',
                text: 'POST https://<你的域名>/v1/chat/completions',
              },
              { type: 'h2', id: 'base-url', text: 'Base URL' },
              {
                type: 'p',
                text: '给 SDK 或客户端填写 Base URL 时，通常写成 https://<你的域名>/v1。程序再去请求 /chat/completions。',
              },
              shot(
                'endpoint.png',
                '首页上圈出的 OpenAI 兼容地址',
                '红圈是首页上的地址示例。请求时换成你正在使用的站点，并保留 /v1。'
              ),
              {
                type: 'note',
                text: '模型广场里的模型不一定都走聊天补全。遇到别的接口类型时，以该模型的实际说明为准。',
              },
            ],
            { label: '打开控制台', target: { type: 'app', to: '/dashboard' } }
          ),
          article(
            'auth',
            '鉴权',
            '在请求头中用 Bearer 方式带上 API 密钥。',
            [
              { type: 'code', text: 'Authorization: Bearer <API_KEY>' },
              { type: 'h2', id: 'key-source', text: '密钥从哪里复制' },
              {
                type: 'ol',
                items: [
                  '登录控制台并打开「API 密钥」。',
                  '使用已经创建好的密钥；新建时立即复制完整内容。',
                  '确认这把密钥的分组包含你要调用的模型。',
                ],
              },
              shot(
                'keys-create.png',
                '控制台 API 密钥页，红圈标出侧栏入口和创建按钮',
                '红圈是「API 密钥」和「创建 API 密钥」。完整密钥只在创建时出现，图里没有真实密钥。'
              ),
              {
                type: 'note',
                text: '示例里的 <API_KEY> 是占位符。不要把真实密钥写进代码仓库、截图或文档。',
              },
            ],
            { label: '打开 API 密钥', target: { type: 'app', to: '/keys' } }
          ),
          article(
            'request',
            '请求示例',
            '下面是一个最小的聊天补全请求，模型名需要换成你实际可用的模型。',
            [
              {
                type: 'code',
                text: [
                  'POST /v1/chat/completions',
                  'Authorization: Bearer <API_KEY>',
                  'Content-Type: application/json',
                  '',
                  '{',
                  '  "model": "<模型名>",',
                  '  "messages": [',
                  '    { "role": "user", "content": "你好" }',
                  '  ]',
                  '}',
                ].join('\n'),
              },
              { type: 'h2', id: 'check-result', text: '如果没有返回' },
              {
                type: 'ul',
                items: [
                  '先看地址有没有重复或漏掉 /v1。',
                  '再看密钥是否完整，以及有没有过期或被停用。',
                  '最后到分组状态确认所选分组可用，并核对模型名。',
                ],
              },
              shot(
                'group-status.png',
                '控制台分组状态页，红圈标出侧栏入口',
                '红圈是「分组状态」。没有正常返回时，可以到这里看所选分组近期有没有成功请求。'
              ),
            ],
            {
              label: '查看常见问题',
              target: { type: 'section', section: 'help' },
            }
          ),
        ],
      },
    ],
  },
  {
    id: 'help',
    title: '常见问题',
    heading: '常见问题',
    description: '先按地址、密钥和分组这三件事排查，不必一上来更换客户端。',
    icon: 'help',
    action: {
      label: '查看分组状态',
      target: { type: 'app', to: '/group-status' },
    },
    groups: [
      {
        id: 'troubleshoot',
        title: '排查',
        articles: [
          article(
            'failed',
            '请求失败',
            '请求没有正常返回时，按下面的顺序缩小范围。',
            [
              { type: 'h2', id: 'order', text: '建议顺序' },
              {
                type: 'ol',
                items: [
                  '确认接口地址使用的是当前站点，并且路径包含 /v1。',
                  '确认 Authorization 头是 Bearer 加上完整密钥。',
                  '确认模型名来自模型广场，且属于这把密钥的分组。',
                  '打开分组状态，看该分组当前是否不可用。',
                ],
              },
              shot(
                'endpoint.png',
                '首页上圈出的 OpenAI 兼容地址',
                '红圈是首页上的接口地址，路径里应包含 /v1。'
              ),
              shot(
                'keys-create.png',
                '控制台 API 密钥页，红圈标出侧栏入口和创建按钮',
                '红圈是核对密钥的位置。图中没有真实密钥。'
              ),
              shot(
                'group-status.png',
                '控制台分组状态页，红圈标出侧栏入口',
                '红圈是控制台里的「分组状态」。'
              ),
              {
                type: 'note',
                text: '这里不列举具体错误码。把客户端或程序返回的原文保留下来，再对照上面四项即可。',
              },
            ],
            {
              label: '打开分组状态',
              target: { type: 'app', to: '/group-status' },
            }
          ),
          article(
            'invalid-key',
            '密钥无效',
            '提示密钥错误、为空或没有权限时，先回到创建密钥的地方核对。',
            [
              { type: 'h2', id: 'check-key', text: '核对这些地方' },
              {
                type: 'ul',
                items: [
                  '复制时没有多出空格或少掉字符。',
                  '客户端里填的是密钥本身，不是密钥名称。',
                  '这把密钥没有被停用或删除。',
                  '请求的模型属于这把密钥选择的分组。',
                ],
              },
              shot(
                'keys-create.png',
                '控制台 API 密钥页，红圈标出侧栏入口和创建按钮',
                '红圈是回到「API 密钥」核对密钥的位置。图中没有真实密钥。'
              ),
            ],
            { label: '打开 API 密钥', target: { type: 'app', to: '/keys' } }
          ),
          article(
            'group-status',
            '查看分组状态',
            '控制台里的分组状态用于查看各个分组当前是否可用。',
            [
              {
                type: 'p',
                text: '某个分组不可用时，选择了该分组的密钥会调用失败。可以换到状态正常、且包含目标模型的分组，或稍后再试。',
              },
              { type: 'h2', id: 'open-status', text: '打开位置' },
              {
                type: 'p',
                text: '登录控制台后，在导航中打开「分组状态」。页面根据最近的调用记录显示各分组是否可用，不是一份固定公告。',
              },
              shot(
                'group-status.png',
                '控制台分组状态页，红圈标出侧栏入口',
                '红圈是侧栏入口。页面数字来自近期调用记录，横线表示这段时间没有请求。'
              ),
            ],
            {
              label: '打开分组状态',
              target: { type: 'app', to: '/group-status' },
            }
          ),
        ],
      },
    ],
  },
]

export function getDocsSection(sectionId: string) {
  return docsSections.find((section) => section.id === sectionId)
}

export function getDocsArticle(sectionId: string, articleSlug: string) {
  const section = getDocsSection(sectionId)
  if (!section) return undefined
  for (const group of section.groups) {
    const found = group.articles.find((item) => item.slug === articleSlug)
    if (found) return { article: found, group, section }
  }
  return undefined
}

export function getDocsNeighbors(sectionId: string, articleSlug: string) {
  const section = getDocsSection(sectionId)
  const articles = section?.groups.flatMap((group) => group.articles) ?? []
  const index = articles.findIndex((item) => item.slug === articleSlug)
  return {
    next: index >= 0 ? articles[index + 1] : undefined,
    previous: index > 0 ? articles[index - 1] : undefined,
  }
}

export type DocsSearchHit = {
  article: DocsArticle
  group: DocsGroup
  section: DocsSection
}

export function searchDocs(query: string) {
  const normalized = query.trim().toLocaleLowerCase()
  if (!normalized) return []
  const hits: DocsSearchHit[] = []
  for (const section of docsSections) {
    for (const group of section.groups) {
      for (const item of group.articles) {
        const haystack = [section.title, group.title, item.title, item.summary]
          .join(' ')
          .toLocaleLowerCase()
        if (haystack.includes(normalized)) {
          hits.push({ article: item, group, section })
        }
      }
    }
  }
  return hits.slice(0, 8)
}
