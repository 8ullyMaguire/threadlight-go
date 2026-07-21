import { goto } from '$app/navigation'
import { page } from '$app/state'
import { profile as currentProfile, type ProfileInfo } from '$lib/app/auth'
import { settings } from '$lib/app/settings.svelte'
import { theme, type ThemeData } from '$lib/app/theme/theme.svelte'
import { fullCommunityName } from '$lib/app/util.svelte'
import type { ResumableItem } from '$lib/feature/legacy/item'
import {
  ArrowRightOnRectangle,
  ArrowTrendingDown,
  ArrowTrendingUp,
  Bookmark,
  ChartBar,
  ChatBubbleLeftRight,
  ChatBubbleOvalLeftEllipsis,
  Clock,
  Cog6Tooth,
  ComputerDesktop,
  Envelope,
  Fire,
  GlobeAlt,
  GlobeAmericas,
  Home,
  Identification,
  Inbox,
  MagnifyingGlass,
  MapPin,
  Moon,
  Newspaper,
  PaintBrush,
  PencilSquare,
  Plus,
  Scale,
  ServerStack,
  ShieldCheck,
  Star,
  Sun,
  Swatch,
  Trophy,
  UserCircle,
  UserGroup,
  ViewColumns,
  type IconSource,
} from '@xylightdev/svelte-hero-icons'
import { SvelteURL } from 'svelte/reactivity'

export interface Group {
  name: string
  actions: Action[]
}

export interface Action {
  name: string
  desc?: string
  handle?: () => void
  href?: string
  shortcut?: string
  icon: string | IconSource
  subActions?: Action[]
  detail?: string
}

export function getGroups(
  resumables: ResumableItem[],
  profile: ProfileInfo,
  profiles: ProfileInfo[],
  td: ThemeData,
  contextual?: Action[],
): Group[] {
  return [
    {
      name: 'Recents',
      actions: resumables.map((r) => ({
        name: r.name,
        icon: r.avatar ?? PencilSquare,
        href: r.url,
      })),
    },
    {
      name: 'On this page',
      actions: contextual ?? [],
    },
    {
      name: 'Main',
      actions: [
        { href: '/', name: 'Home', icon: Home, shortcut: 'h' },
        {
          href: '/communities',
          name: 'Communities',
          icon: GlobeAlt,
        },
        ...(currentProfile.isMod()
          ? [
              {
                href: '/moderation',
                name: 'Moderation',
                icon: ShieldCheck,
              },
            ]
          : []),
        ...(currentProfile.isAdmin
          ? [
              {
                href: '/admin',
                name: 'Admin',
                icon: ServerStack,
              },
            ]
          : []),
      ],
    },
    {
      name: 'Feeds',
      actions: [
        {
          name: 'Location',
          icon: GlobeAmericas,
          subActions: [
            {
              name: 'All',
              icon: GlobeAmericas,
              href: '/?type=All',
            },
            {
              name: 'Local',
              icon: MapPin,
              href: '/?type=Local',
            },
            {
              name: 'Subscriptions',
              icon: Newspaper,
              href: '/?type=Subscribed',
            },
          ],
        },
        {
          name: 'Sort',
          icon: ChartBar,
          subActions: [
            {
              name: 'Top',
              icon: Trophy,
              subActions: [
                {
                  name: 'All time',
                  icon: ChartBar,
                  href: '/?sort=TopAll',
                },
                {
                  name: '9 months',
                  icon: ChartBar,
                  href: '/?sort=TopNineMonths',
                },
                {
                  name: '6 months',
                  icon: ChartBar,
                  href: '/?sort=TopSixMonths',
                },
                {
                  name: '3 months',
                  icon: ChartBar,
                  href: '/?sort=TopThreeMonths',
                },
                {
                  name: 'Past month',
                  icon: ChartBar,
                  href: '/?sort=TopMonth',
                },
                {
                  name: 'Past week',
                  icon: ChartBar,
                  href: '/?sort=TopWeek',
                },
                {
                  name: 'Past day',
                  icon: ChartBar,
                  href: '/?sort=TopDay',
                },
                {
                  name: '6 hours',
                  icon: ChartBar,
                  href: '/?sort=TopSixHour',
                },
                {
                  name: 'Past hour',
                  icon: ChartBar,
                  href: '/?sort=TopHour',
                },
              ],
            },
            {
              name: 'Active',
              icon: ArrowTrendingUp,
              href: '/?sort=Active',
            },
            {
              name: 'Hot',
              icon: Fire,
              href: '/?sort=Hot',
            },
            {
              name: 'Scaled',
              icon: Scale,
              href: '/?sort=Scaled',
            },
            {
              name: 'New',
              icon: Star,
              href: '/?sort=New',
            },
            {
              name: 'Old',
              icon: Clock,
              href: '/?sort=Old',
            },
            {
              name: 'Controversial',
              icon: ArrowTrendingDown,
              href: '/?sort=Controversial',
            },
            {
              name: 'Most Comments',
              icon: ChatBubbleOvalLeftEllipsis,
              href: '/?sort=MostComments',
            },
            {
              name: 'New Replies',
              icon: ChatBubbleLeftRight,
              href: '/?sort=NewComments',
            },
          ],
        },
      ],
    },
    {
      name: 'Profile',
      actions: profile.jwt
        ? [
            {
              href: '/profile/user',
              name: 'Profile',
              icon: UserCircle,
            },
            {
              href: '/inbox',
              name: 'Inbox',
              icon: Inbox,
              shortcut: 'i',
            },
            {
              href: '/saved',
              name: 'Saved',
              icon: Bookmark,
            },
            {
              href: '/accounts',
              name: 'Accounts',
              icon: UserGroup,
            },
            {
              href: '/login',
              name: 'Log in',
              icon: ArrowRightOnRectangle,
            },
            {
              href: '/signup',
              name: 'Sign up',
              icon: Identification,
            },
            {
              href: '/login/guest',
              name: 'Add guest',
              icon: Plus,
            },
            {
              href: '/inbox/messages',
              name: 'Messages',
              icon: Envelope,
            },
          ]
        : [
            {
              href: '/login',
              name: 'Log in',
              icon: ArrowRightOnRectangle,
            },
            {
              href: '/signup',
              name: 'Sign up',
              icon: Identification,
            },
            {
              href: '/login/guest',
              name: 'Add guest',
              icon: Plus,
            },
            {
              href: '/accounts',
              name: 'Accounts',
              icon: UserGroup,
            },
          ],
    },
    {
      name: 'Accounts',
      actions: profiles.map((p) => ({
        name: p.username ?? 'Guest',
        icon: p.avatar ?? UserCircle,
        detail: p.instance,
        handle: async () => {
          if (profile.id != p.id) {
            currentProfile.meta.profile = p.id
          }

          await goto(page.url, {
            invalidateAll: true,
          })
        },
      })),
    },
    {
      name: 'App',
      actions: [
        {
          href: '/settings',
          name: 'Settings',
          icon: Cog6Tooth,
        },
        {
          name: 'Set view',
          icon: ViewColumns,
          subActions: [
            {
              name: `Set view to **${'Compact'}**`,
              icon: ViewColumns,
              handle: () => (settings.view = 'compact'),
            },
            {
              name: `Set view to **${'Cozy'}**`,
              icon: ViewColumns,
              handle: () => (settings.view = 'cozy'),
            },
          ],
        },
        {
          name: 'Set color scheme',
          icon: PaintBrush,
          subActions: [
            {
              name: `Set color scheme to **${'System'}**`,
              handle: () => (theme.colorScheme = 'system'),
              icon: ComputerDesktop,
            },
            {
              name: `Set color scheme to **${'Light'}**`,
              handle: () => (theme.colorScheme = 'light'),
              icon: Sun,
            },
            {
              name: `Set color scheme to **${'Dark'}**`,
              handle: () => (theme.colorScheme = 'dark'),
              icon: Moon,
            },
          ],
        },
        {
          name: 'Set theme',
          icon: Swatch,
          subActions: td.themes.map((th) => ({
            name: `Set theme to **${th.name}**`,
            icon: Swatch,
            handle: () => (theme.data.currentTheme = th.id),
          })),
        },
      ],
    },
    {
      name: 'Content',
      actions: [
        {
          href: '/create/post',
          name: 'Create post',
          icon: PencilSquare,
        },
      ],
    },
    {
      name: 'Subscriptions',
      actions:
        profile.user?.follows.map((f) => ({
          icon: f.community.icon ?? Newspaper,
          name: f.community.title,
          href: `/c/${fullCommunityName(
            f.community.name,
            f.community.actor_id,
          )}`,
          detail: new SvelteURL(f.community.actor_id).hostname,
        })) ?? [],
    },
  ]
}

// this should really be rewritten, but a temporary change
// to make sure we don't recalculate ALL of that per keypress
export function dynamicActions(query: string): Group {
  const actions = [
    {
      name: `Search for **${query.trim()}**`,
      icon: MagnifyingGlass,
      href: `/search?q=${query}`,
    },
    {
      name: `Communities with **${query.trim()}**`,
      icon: Newspaper,
      href: `/explore/communities?q=${query}`,
    },
  ]

  return {
    name: 'Other results',
    actions: actions,
  }
}
