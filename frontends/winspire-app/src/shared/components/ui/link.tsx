/**
 * TODO: Update this component to use your client-side framework's link
 * component. We've provided examples of how to do this for Next.js, Remix, and
 * Inertia.js in the Catalyst documentation:
 *
 * https://catalyst.tailwindui.com/docs#client-side-router-integration
 */

import * as Headless from '@headlessui/react'
import React, { forwardRef } from 'react'
import { Link as RouterLink } from 'react-router-dom'

type LinkProps = { href: string } & Omit<React.ComponentPropsWithoutRef<typeof RouterLink>, 'to'>

export const Link = forwardRef(function Link({ href, ...props }: LinkProps, ref: React.ForwardedRef<HTMLAnchorElement>) {
  return (
    <Headless.DataInteractive>
      <RouterLink {...props} to={href} ref={ref} />
    </Headless.DataInteractive>
  )
})
