<script>
	import { t } from '$lib/locale';
	import Button from '$components/ui/button/button.svelte';
	import * as Card from '$components/ui/card';

	import ArrowRightIcon from '@lucide/svelte/icons/arrow-right';
	import Footer from '$components/nav/Footer.svelte';
	import BasicHeader from '$components/nav/BasicHeader.svelte';

	let { data } = $props();
</script>

<div class="grid min-h-screen grid-rows-[auto_1fr_auto]">
	<BasicHeader />

	<main class="my-16 px-4 md:px-8">
		<div
			class="mx-auto flex max-w-screen-2xl flex-col items-center justify-around space-y-6 text-center"
		>
			<div class="leading-relaxed">
				<h1 class="text-primary text-5xl font-bold md:text-8xl">
					{$t('landing.title')}
				</h1>
				<h2 class="text-5xl font-bold md:text-8xl">
					{$t('landing.hero')}
				</h2>
			</div>

			<a
				href="/register"
				class="border-primary hover:bg-primary group flex w-full items-center justify-between gap-4 rounded-full border p-2 transition-colors duration-200 md:w-min"
			>
				<p class="ml-4 text-xl font-medium md:text-2xl">
					<nobr>{$t('landing.join-button')}</nobr>
				</p>

				<div
					class="bg-primary group-hover:bg-primary-foreground flex items-center justify-center rounded-full px-8 py-2 transition-colors duration-200"
				>
					<ArrowRightIcon
						class="stroke-primary-foreground group-hover:stroke-primary transition-colors duration-200"
					/>
				</div>
			</a>
		</div>

		<img src="/landing/hero.webp" alt="company" class="mx-auto my-16 w-full max-w-screen-2xl" />

		<div class="mx-auto grid max-w-screen-2xl grid-cols-1 gap-6 md:grid-cols-3">
			<Card.Root>
				<Card.Header>
					<Card.Title>Anonymous e-Sim internet access for 150+ countries</Card.Title>
				</Card.Header>
				<Card.Content>
					<img src="/landing/card_1.webp" alt="company" class="h-36 w-full object-contain" />
				</Card.Content>
				<Card.Footer>
					<Button class="w-full" href="/register">Get Company Now</Button>
				</Card.Footer>
			</Card.Root>
			<Card.Root>
				<Card.Header>
					<Card.Title>Farm $COMPANY_TOKEN token - the future of global internet</Card.Title>
				</Card.Header>
				<Card.Content>
					<img src="/landing/card_2.webp" alt="company" class="h-36 w-full object-contain" />
				</Card.Content>
				<Card.Footer>
					<Button class="w-full" href="/register">Start farming</Button>
				</Card.Footer>
			</Card.Root>
			<Card.Root>
				<Card.Header>
					<Card.Title>Unique buyback model with up to 100% revenue per month</Card.Title>
				</Card.Header>
				<Card.Content>
					<img src="/landing/card_3.webp" alt="company" class="h-36 w-full object-contain" />
				</Card.Content>
				<Card.Footer>
					<Button class="w-full" href="/register">Start earning</Button>
				</Card.Footer>
			</Card.Root>
		</div>
	</main>

	<Footer {data} />
</div>
