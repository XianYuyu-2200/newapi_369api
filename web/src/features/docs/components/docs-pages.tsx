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

import { useEffect, useState } from 'react'
import { Link } from '@tanstack/react-router'
import {
  ArrowRight,
  BookOpen,
  ChevronDown,
  CircleHelp,
  Code2,
  Info,
  Menu,
  Terminal,
  X,
  type LucideIcon,
} from 'lucide-react'

import { cn } from '@/lib/utils'

import { useDocsChrome } from '../hooks/use-docs-chrome'
import { docsAmberClass, docsPrimaryClass, docsSecondaryClass } from '../lib/docs-classes'
import {
  docsHome,
  docsSections,
  getDocsArticle,
  getDocsNeighbors,
  getDocsSection,
  type DocsArticle,
  type DocsBlock,
  type DocsGroup,
  type DocsIcon,
  type DocsSection,
} from '../lib/content'
import {
  DocsAnchor,
  DocsChromeProvider,
  DocsExternalIcon,
  DocsLogo,
  DocsSearch,
  DocsThemeButton,
} from './docs-chrome'

const sectionIcons: Record<DocsIcon, LucideIcon> = {
  book: BookOpen,
  code: Code2,
  help: CircleHelp,
  terminal: Terminal,
}

function useDocsTitle(title: string) {
  useEffect(() => {
    const previous = document.title
    document.title = title
    return () => {
      document.title = previous
    }
  }, [title])
}

function DocsMark({ className }: { className?: string }) {
  return (
    <svg viewBox='0 0 240 240' className={className} aria-hidden='true'>
      <defs>
        <linearGradient id='docs-mark-gradient' x1='30' y1='20' x2='210' y2='220'>
          <stop offset='0' stopColor='#fb7185' />
          <stop offset='0.48' stopColor='#a78bfa' />
          <stop offset='1' stopColor='#38bdf8' />
        </linearGradient>
      </defs>
      <path
        d='M146 24c46 8 74 52 64 98-10 48-54 70-92 90-34 18-74 34-96 12-24-24 0-66 20-102C62 78 96 34 146 24z'
        fill='url(#docs-mark-gradient)'
      />
      <path
        d='M70 146c24-34 66-44 104-28 20 8 38 28 32 48-8 30-52 40-88 30-28-8-46-24-48-50z'
        fill='#60a5fa'
        opacity='0.9'
      />
    </svg>
  )
}

function PortalHeader() {
  const [sectionsOpen, setSectionsOpen] = useState(false)
  return (
    <header className='sticky top-0 z-40 border-b border-[var(--docs-line)] bg-[var(--docs-header)] backdrop-blur'>
      <div className='mx-auto flex h-16 max-w-6xl items-center gap-3 px-4 md:px-6'>
        <button
          type='button'
          className='grid size-10 place-items-center rounded-xl border border-[var(--docs-line)] md:hidden'
          aria-label={sectionsOpen ? '关闭栏目' : '打开栏目'}
          aria-expanded={sectionsOpen}
          onClick={() => setSectionsOpen((open) => !open)}
        >
          {sectionsOpen ? (
            <X className='size-4' aria-hidden='true' />
          ) : (
            <Menu className='size-4' aria-hidden='true' />
          )}
        </button>
        <DocsLogo />
        <DocsSearch className='mx-auto hidden min-w-0 max-w-xl flex-1 sm:block' />
        <div className='ml-auto flex items-center gap-2'>
          <DocsAnchor target={{ type: 'app', to: '/' }} className={docsSecondaryClass}>
            主站
            <DocsExternalIcon />
          </DocsAnchor>
          <DocsThemeButton />
        </div>
      </div>
      <DocsSearch className='px-4 pb-3 sm:hidden' />
      {sectionsOpen ? (
        <nav className='border-t border-[var(--docs-line)] px-4 py-3 md:hidden'>
          <ul className='space-y-1'>
            {docsSections.map((section) => (
              <li key={section.id}>
                <DocsAnchor
                  target={{ type: 'section', section: section.id }}
                  className='block rounded-lg px-3 py-2 text-sm font-medium hover:bg-[var(--docs-hover)]'
                  onClick={() => setSectionsOpen(false)}
                >
                  {section.title}
                </DocsAnchor>
              </li>
            ))}
          </ul>
        </nav>
      ) : null}
    </header>
  )
}

function ArticleSidebar({
  articleSlug,
  className,
  onNavigate,
  section,
}: {
  articleSlug: string
  className?: string
  onNavigate?: () => void
  section: DocsSection
}) {
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({})

  useEffect(() => {
    setCollapsed({})
  }, [articleSlug, section.id])

  const isOpen = (group: DocsGroup) => {
    if (group.id in collapsed) return !collapsed[group.id]
    return group.articles.some((item) => item.slug === articleSlug)
  }

  return (
    <aside
      className={cn(
        'flex flex-col bg-[var(--docs-sidebar)] text-[var(--docs-text)]',
        className
      )}
    >
      <div className='flex items-center justify-between gap-3 px-4 pt-4'>
        <DocsLogo />
        {onNavigate ? (
          <button
            type='button'
            className='grid size-9 place-items-center rounded-lg hover:bg-[var(--docs-hover)] lg:hidden'
            aria-label='关闭目录'
            onClick={onNavigate}
          >
            <X className='size-4' aria-hidden='true' />
          </button>
        ) : null}
      </div>
      <DocsSearch className='px-4 py-4' />
      <nav aria-label='文档目录' className='min-h-0 flex-1 overflow-y-auto px-3 pb-6'>
        {section.groups.map((group) => {
          const open = isOpen(group)
          return (
            <div key={group.id} className='mb-2'>
              <button
                type='button'
                className='flex w-full items-center justify-between rounded-lg px-2 py-2 text-left text-sm text-[var(--docs-muted)] hover:bg-[var(--docs-hover)]'
                aria-expanded={open}
                onClick={() =>
                  setCollapsed((current) => ({ ...current, [group.id]: open }))
                }
              >
                <span>{group.title}</span>
                <ChevronDown
                  className={cn('size-4 transition', !open && '-rotate-90')}
                  aria-hidden='true'
                />
              </button>
              {open ? (
                <div className='mt-1 space-y-1'>
                  {group.articles.map((item) => {
                    const active = item.slug === articleSlug
                    return (
                      <DocsAnchor
                        key={item.slug}
                        target={{
                          type: 'article',
                          section: section.id,
                          article: item.slug,
                        }}
                        onClick={onNavigate}
                        className={cn(
                          'block rounded-lg px-3 py-2 text-sm',
                          active
                            ? 'bg-[var(--docs-accent-soft)] font-medium text-[var(--docs-accent)]'
                            : 'hover:bg-[var(--docs-hover)]'
                        )}
                      >
                        {item.title}
                      </DocsAnchor>
                    )
                  })}
                </div>
              ) : null}
            </div>
          )
        })}
      </nav>
    </aside>
  )
}

function ArticleTopBar({ sectionId }: { sectionId: string }) {
  const { menuOpen, setMenuOpen } = useDocsChrome()
  return (
    <div className='sticky top-0 z-30 flex items-center gap-3 border-b border-[var(--docs-line)] bg-[var(--docs-header)] px-4 backdrop-blur lg:px-8'>
      <button
        type='button'
        className='grid size-10 shrink-0 place-items-center rounded-xl border border-[var(--docs-line)] lg:hidden'
        aria-label={menuOpen ? '关闭目录' : '打开目录'}
        aria-expanded={menuOpen}
        onClick={() => setMenuOpen(!menuOpen)}
      >
        {menuOpen ? (
          <X className='size-4' aria-hidden='true' />
        ) : (
          <Menu className='size-4' aria-hidden='true' />
        )}
      </button>
      <nav className='flex min-w-0 flex-1 items-center gap-5 overflow-x-auto text-sm font-semibold'>
        {docsSections.map((section) => {
          const active = section.id === sectionId
          return (
            <DocsAnchor
              key={section.id}
              target={{ type: 'section', section: section.id }}
              className={cn(
                'shrink-0 border-b-2 py-4',
                active
                  ? 'border-[var(--docs-accent)] text-[var(--docs-accent)]'
                  : 'border-transparent text-[var(--docs-text)]'
              )}
            >
              {section.title}
            </DocsAnchor>
          )
        })}
      </nav>
      <div className='flex shrink-0 items-center gap-2'>
        <DocsAnchor
          target={{ type: 'app', to: '/' }}
          className={cn(docsSecondaryClass, 'hidden sm:inline-flex')}
        >
          主站
          <DocsExternalIcon />
        </DocsAnchor>
        <DocsThemeButton />
        <DocsAnchor target={{ type: 'app', to: '/dashboard' }} className={docsPrimaryClass}>
          控制台
        </DocsAnchor>
      </div>
    </div>
  )
}


function blockKey(block: DocsBlock) {
  if (block.type === 'h2') return block.id
  if (block.type === 'code') return `code:${block.text}`
  if (block.type === 'image') return `image:${block.src}`
  if (block.type === 'p' || block.type === 'note') return `${block.type}:${block.text}`
  return `${block.type}:${block.items.join('|')}`
}

function ArticleBody({ article }: { article: DocsArticle }) {
  return (
    <div className='max-w-3xl'>
      {article.blocks.map((block) => {
        const key = blockKey(block)
        if (block.type === 'p') {
          return (
            <p key={key} className='mt-4 text-base leading-7 text-[var(--docs-muted)]'>
              {block.text}
            </p>
          )
        }
        if (block.type === 'h2') {
          return (
            <h2
              key={key}
              id={block.id}
              className='mt-10 scroll-mt-24 text-2xl font-semibold tracking-tight'
            >
              {block.text}
            </h2>
          )
        }
        if (block.type === 'note') {
          return (
            <div
              key={key}
              className='mt-4 rounded-xl border border-[var(--docs-line)] bg-[var(--docs-card)] p-4'
            >
              <p className='flex items-center gap-2 text-sm font-semibold text-[var(--docs-accent)]'>
                <Info className='size-4' aria-hidden='true' />
                说明
              </p>
              <p className='mt-2 text-sm leading-6 text-[var(--docs-muted)]'>{block.text}</p>
            </div>
          )
        }
        if (block.type === 'code') {
          return (
            <pre
              key={key}
              className='mt-4 overflow-x-auto rounded-xl bg-[var(--docs-code)] p-4 text-sm leading-6 text-white'
            >
              {block.text}
            </pre>
          )
        }
        if (block.type === 'image') {
          return (
            <figure key={key} className='mt-4'>
              <img
                src={block.src}
                alt={block.alt}
                className='h-auto max-w-full rounded-xl border border-[var(--docs-line)]'
              />
              <figcaption className='mt-2 text-sm leading-6 text-[var(--docs-muted)]'>
                {block.caption}
              </figcaption>
            </figure>
          )
        }
        const ListTag = block.type === 'ol' ? 'ol' : 'ul'
        return (
          <ListTag
            key={key}
            className={cn(
              'mt-4 space-y-2 pl-5 text-base leading-7 text-[var(--docs-muted)]',
              block.type === 'ol' ? 'list-decimal' : 'list-disc'
            )}
          >
            {block.items.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ListTag>
        )
      })}
    </div>
  )
}

function ArticleToc({ article }: { article: DocsArticle }) {
  const headings = article.blocks.flatMap((block) =>
    block.type === 'h2' ? [block] : []
  )
  const headingKey = headings.map((heading) => heading.id).join('|')
  const [active, setActive] = useState(headings[0]?.id ?? '')

  useEffect(() => {
    const ids = headingKey.split('|').filter(Boolean)
    const elements = ids.flatMap((id) => {
      const element = document.querySelector(`[id="${CSS.escape(id)}"]`)
      return element ? [element] : []
    })
    if (elements.length === 0) return
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((entry) => entry.isIntersecting)
          .sort((left, right) => right.intersectionRatio - left.intersectionRatio)[0]
        if (visible?.target.id) setActive(visible.target.id)
      },
      { rootMargin: '-96px 0px -55% 0px', threshold: [0.2, 1] }
    )
    elements.forEach((element) => observer.observe(element))
    return () => observer.disconnect()
  }, [headingKey])

  if (headings.length === 0) return null

  return (
    <nav aria-label='本页内容' className='sticky top-20 hidden h-fit py-8 pr-6 xl:block'>
      <p className='text-sm font-semibold'>本页内容</p>
      <ul className='mt-3 border-l border-[var(--docs-line)]'>
        {headings.map((heading) => (
          <li key={heading.id}>
            <a
              href={`#${heading.id}`}
              className={cn(
                'block border-l-2 px-3 py-1.5 text-sm',
                active === heading.id
                  ? 'border-[var(--docs-accent)] bg-[var(--docs-accent-soft)] text-[var(--docs-accent)]'
                  : 'border-transparent text-[var(--docs-muted)] hover:text-[var(--docs-text)]'
              )}
            >
              {heading.text}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  )
}

function MissingDocs({ title }: { title: string }) {
  useDocsTitle(title)
  return (
    <DocsChromeProvider>
      <PortalHeader />
      <main id='docs-content' className='mx-auto max-w-3xl px-6 py-20'>
        <h1 className='text-4xl font-bold'>{title}</h1>
        <p className='mt-4 text-[var(--docs-muted)]'>这篇内容还不存在，或链接写错了。</p>
        <Link to='/docs' className={cn(docsPrimaryClass, 'mt-6')}>
          返回文档首页
        </Link>
      </main>
    </DocsChromeProvider>
  )
}

export function DocsHome() {
  useDocsTitle('369api 文档')
  return (
    <DocsChromeProvider>
      <PortalHeader />
      <main id='docs-content' className='mx-auto max-w-6xl px-4 py-6 md:px-6 md:py-8'>
        <section className='docs-hero overflow-hidden rounded-2xl border border-[var(--docs-line)]'>
          <div className='grid items-center gap-6 px-6 py-10 md:grid-cols-[minmax(0,1fr)_240px] md:px-10 md:py-12'>
            <div>
              <p className='inline-flex rounded-full border border-[var(--docs-accent)]/40 px-3 py-1 text-xs font-medium text-[var(--docs-accent)]'>
                {docsHome.eyebrow}
              </p>
              <h1 className='mt-5 text-4xl font-bold tracking-tight md:text-5xl'>
                {docsHome.title}
              </h1>
              <p className='mt-4 text-xl font-semibold md:text-2xl'>{docsHome.headline}</p>
              <p className='mt-3 max-w-2xl text-sm leading-6 text-[var(--docs-muted)] md:text-base'>
                {docsHome.description}
              </p>
              <div className='mt-6 flex flex-wrap gap-3'>
                <DocsAnchor
                  target={{ type: 'article', section: 'start', article: 'codex-install' }}
                  className={docsPrimaryClass}
                >
                  从第一步开始
                  <ArrowRight className='size-4' aria-hidden='true' />
                </DocsAnchor>
                <DocsAnchor target={{ type: 'app', to: '/dashboard' }} className={docsSecondaryClass}>
                  进入控制台
                  <DocsExternalIcon />
                </DocsAnchor>
                <DocsAnchor target={{ type: 'app', to: '/pricing' }} className={docsAmberClass}>
                  模型广场
                  <DocsExternalIcon />
                </DocsAnchor>
              </div>
            </div>
            <DocsMark className='mx-auto hidden h-52 w-52 md:block' />
          </div>
        </section>
        <section
          id='docs-sections'
          className='mt-6 grid gap-4 md:grid-cols-2'
          aria-label='文档栏目'
        >
          {docsSections.map((section) => {
            const Icon = sectionIcons[section.icon]
            return (
              <DocsAnchor
                key={section.id}
                target={{ type: 'section', section: section.id }}
                className='flex items-center gap-4 rounded-xl border border-[var(--docs-line)] bg-[var(--docs-card)] p-5 hover:bg-[var(--docs-hover)]'
              >
                <span className='grid size-11 shrink-0 place-items-center rounded-xl bg-[var(--docs-accent-soft)] text-[var(--docs-accent)]'>
                  <Icon className='size-5' aria-hidden='true' />
                </span>
                <span className='min-w-0 flex-1'>
                  <span className='block text-base font-semibold'>{section.title}</span>
                  <span className='mt-1 block text-sm leading-6 text-[var(--docs-muted)]'>
                    {section.description}
                  </span>
                </span>
                <ArrowRight className='size-4 shrink-0 text-[var(--docs-faint)]' aria-hidden='true' />
              </DocsAnchor>
            )
          })}
        </section>
      </main>
    </DocsChromeProvider>
  )
}

export function DocsSectionPage({ sectionId }: { sectionId: string }) {
  const section = getDocsSection(sectionId)
  useDocsTitle(section ? `${section.heading} | 369api 文档` : '没有这个栏目')
  if (!section) return <MissingDocs title='没有这个栏目' />
  return (
    <DocsChromeProvider>
      <PortalHeader />
      <main id='docs-content' className='mx-auto max-w-5xl px-4 py-10 md:px-6 md:py-14'>
        <h1 className='text-4xl font-bold tracking-tight'>{section.heading}</h1>
        <p className='mt-4 max-w-2xl text-base leading-7 text-[var(--docs-muted)]'>
          {section.description}
        </p>
        {section.action ? (
          <DocsAnchor target={section.action.target} className={cn(docsPrimaryClass, 'mt-6')}>
            {section.action.label}
            <ArrowRight className='size-4' aria-hidden='true' />
          </DocsAnchor>
        ) : null}
        {section.groups.map((group) => (
          <section key={group.id} className='mt-12'>
            <h2 className='text-2xl font-semibold'>{group.title}</h2>
            <div className='mt-5 grid gap-4 md:grid-cols-2'>
              {group.articles.map((item) => {
                const Icon = sectionIcons[section.icon]
                return (
                  <DocsAnchor
                    key={item.slug}
                    target={{ type: 'article', section: section.id, article: item.slug }}
                    className='rounded-xl border border-[var(--docs-line)] bg-[var(--docs-card)] p-5 hover:bg-[var(--docs-hover)]'
                  >
                    <Icon className='size-5 text-[var(--docs-text)]' aria-hidden='true' />
                    <span className='mt-4 block font-semibold'>{item.title}</span>
                    <span className='mt-2 block text-sm leading-6 text-[var(--docs-muted)]'>
                      {item.summary}
                    </span>
                  </DocsAnchor>
                )
              })}
            </div>
          </section>
        ))}
      </main>
    </DocsChromeProvider>
  )
}

export function DocsArticlePage({
  articleSlug,
  sectionId,
}: {
  articleSlug: string
  sectionId: string
}) {
  const found = getDocsArticle(sectionId, articleSlug)
  useDocsTitle(found ? `${found.article.title} | 369api 文档` : '没有这篇文章')
  if (!found) return <MissingDocs title='没有这篇文章' />
  return (
    <DocsChromeProvider>
      <DocsArticleLayout found={found} />
    </DocsChromeProvider>
  )
}

function DocsArticleLayout({
  found,
}: {
  found: NonNullable<ReturnType<typeof getDocsArticle>>
}) {
  const { menuOpen, setMenuOpen } = useDocsChrome()
  const { article, group, section } = found
  const neighbors = getDocsNeighbors(section.id, article.slug)
  return (
      <div className='lg:grid lg:grid-cols-[280px_minmax(0,1fr)]'>
        <ArticleSidebar
          section={section}
          articleSlug={article.slug}
          className='sticky top-0 hidden h-svh lg:flex'
        />
        {menuOpen ? (
          <div className='fixed inset-0 z-50 lg:hidden'>
            <button
              type='button'
              className='absolute inset-0 bg-black/60'
              aria-label='关闭目录'
              onClick={() => setMenuOpen(false)}
            />
            <ArticleSidebar
              section={section}
              articleSlug={article.slug}
              onNavigate={() => setMenuOpen(false)}
              className='relative h-full w-[min(100%,320px)]'
            />
          </div>
        ) : null}
        <div className='min-w-0'>
          <ArticleTopBar sectionId={section.id} />
          <div className='xl:grid xl:grid-cols-[minmax(0,1fr)_240px]'>
            <article id='docs-content' className='px-5 py-8 md:px-10'>
              <p className='text-sm text-[var(--docs-accent)]'>
                <DocsAnchor
                  target={{ type: 'section', section: section.id }}
                  className='hover:underline'
                >
                  {group.title}
                </DocsAnchor>
                <span className='px-2 text-[var(--docs-faint)]'>/</span>
                {article.title}
              </p>
              <h1 className='mt-3 text-4xl font-bold tracking-tight'>{article.title}</h1>
              <p className='mt-4 text-base leading-7 text-[var(--docs-muted)]'>
                {article.summary}
              </p>
              {article.action ? (
                <DocsAnchor target={article.action.target} className={cn(docsAmberClass, 'mt-6')}>
                  {article.action.label}
                  <DocsExternalIcon />
                </DocsAnchor>
              ) : null}
              <ArticleBody article={article} />
              <div className='mt-12 grid gap-3 sm:grid-cols-2'>
                {neighbors.previous ? (
                  <DocsAnchor
                    target={{
                      type: 'article',
                      section: section.id,
                      article: neighbors.previous.slug,
                    }}
                    className='rounded-xl border border-[var(--docs-line)] bg-[var(--docs-card)] p-4 hover:bg-[var(--docs-hover)]'
                  >
                    <span className='text-xs text-[var(--docs-faint)]'>上一篇</span>
                    <span className='mt-1 block font-semibold'>{neighbors.previous.title}</span>
                  </DocsAnchor>
                ) : (
                  <span />
                )}
                {neighbors.next ? (
                  <DocsAnchor
                    target={{
                      type: 'article',
                      section: section.id,
                      article: neighbors.next.slug,
                    }}
                    className='rounded-xl border border-[var(--docs-line)] bg-[var(--docs-card)] p-4 text-right hover:bg-[var(--docs-hover)]'
                  >
                    <span className='text-xs text-[var(--docs-faint)]'>下一篇</span>
                    <span className='mt-1 block font-semibold'>{neighbors.next.title}</span>
                  </DocsAnchor>
                ) : null}
              </div>
            </article>
            <ArticleToc article={article} />
          </div>
        </div>
      </div>
  )
}

