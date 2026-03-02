<script lang="ts">
	import * as Card from '$components/ui/card';
	import { Button } from '$components/ui/button';
</script>

<div class="container mx-auto space-y-6 px-4 py-8 md:space-y-8">
	<h1 class="text-3xl font-bold">Company opportunities</h1>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<Card.Root>
			<Card.Header>
				<Card.Title>Replenish your balance and get anonymous eSim access</Card.Title>
			</Card.Header>
			<Card.Content>
				<img src="/home/slide_1.webp" alt="eSim" />
				<p>
					Deposit USDT or USDC on Binance Smart Chain and choose one of our e-Sim options for your
					needs.
				</p>
			</Card.Content>
			<Card.Footer>
				<Button class="w-full" href="/app/esim">Deposit and get eSim</Button>
			</Card.Footer>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title>Farm $COMPANY_TOKEN token with Farming towers</Card.Title>
			</Card.Header>
			<Card.Content>
				<img src="/home/slide_2.webp" alt="eSim" />
				<p>
					Get one of the Farming towers to farm $COMPANY_TOKEN tokens passively. Use the towers in future
					updates to claim your part of global connection map.
				</p>
			</Card.Content>
			<Card.Footer>
				<Button class="w-full" href="/app/farm">Start farming $COMPANY_TOKEN</Button>
			</Card.Footer>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title>Earn up to 100% / month</Card.Title>
			</Card.Header>
			<Card.Content>
				<img src="/home/slide_3.webp" alt="eSim" />
				<p>
					Hold your $COMPANY_TOKEN tokens or use our unique price-based buyback model to sell the tokens
					with a lowest price with up to 100% monthly profit.
				</p>
			</Card.Content>
			<Card.Footer>
				<Button class="w-full" href="/app/farm">Start earning with $COMPANY_TOKEN</Button>
			</Card.Footer>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title>Invite your friends and earn even more</Card.Title>
			</Card.Header>
			<Card.Content>
				<img src="/home/slide_4.webp" alt="eSim" />
				<p>
					Use your referral link and invite your community and earn every time they start farming
					$COMPANY_TOKEN tokens in infinite amount of lines with a business difference referral model.
				</p>
			</Card.Content>
			<Card.Footer>
				<Button class="w-full" href="/app/referral">Invite friends</Button>
			</Card.Footer>
		</Card.Root>
	</div>

	<h1 class="text-3xl font-bold">How to start with Company Network</h1>

	<div class="grid grid-cols-2 md:grid-cols-3 gap-4">
		<div>
			<div class="bg-primary/20 h-[30px] w-full rounded-full p-1">
				<div class="bg-primary h-[22px] w-[22px] rounded-full text-center">1</div>
			</div>
			<p class="border-r text-xs mt-2 mr-4 pr-4">
				<b><u>Replenish</u></b> your balance with USDT or USDC on Binance Smart Chain
			</p>
		</div>
		<div>
			<div class="bg-primary/20 h-[30px] w-full rounded-full p-1">
				<div class="bg-primary h-[22px] w-[22px] rounded-full text-center">2</div>
			</div>
			<p class="border-r text-xs mt-2 mr-4 pr-4">
				<b><u>Activate</u></b> your eSIM with your desired plan
			</p>
		</div>
		<div>
			<div class="bg-primary/20 h-[30px] w-full rounded-full p-1">
				<div class="bg-primary h-[22px] w-[22px] rounded-full text-center">3</div>
			</div>
			<p class="border-r text-xs mt-2 mr-4 pr-4">
				<b><u>Purchase</u></b> one of our farming towers and earn from 70% to 100% monthly
			</p>
		</div>
		<div>
			<div class="bg-primary/20 h-[30px] w-full rounded-full p-1">
				<div class="bg-primary h-[22px] w-[22px] rounded-full text-center">4</div>
			</div>
			<p class="border-r text-xs mt-2 mr-4 pr-4">
				<b><u>Copy your referral link</u></b> and invite your friends and earn up to 17% referral compensation
				in infinite depth
			</p>
		</div>
		<div>
			<div class="bg-primary/20 h-[30px] w-full rounded-full p-1">
				<div class="bg-primary h-[22px] w-[22px] rounded-full text-center">5</div>
			</div>
			<p class="border-r text-xs mt-2 mr-4 pr-4">
				You will receive $COMPANY_TOKEN tokens everyday. Hold them or sell it with our unique 
				<b><u>Bonding Curve Buyback model.</u></b> 
				Important: Claiming the tokens may include fees that depends on a token growth rate. Initially the claim fee is 99% and it reduces to 0% following the token growth.
			</p>
		</div>
		<div>
			<div class="bg-primary/20 h-[30px] w-full rounded-full p-1">
				<div class="bg-primary h-[22px] w-[22px] rounded-full text-center">6</div>
			</div>
			<p class="border-r text-xs mt-2 mr-4 pr-4">
				Achieve leadership career status and
				<b><u>get rewarded. </u></b> 
				Only 33% of each structure turnover is counted to reaching the new Status.
			</p>
		</div>
	</div>

	<h1 class="text-3xl font-bold">Get prepared for Stage 2 - worldwide connection strategy game</h1>

	<div class="grid grid-cols-1 grid-rows-4 md:grid-cols-3 md:grid-rows-3 gap-6">
		<div 
			class="col-span-1 row-span-1 md:row-span-2 md:col-span-3 bg-cover bg-center rounded-xl" 
			style="background-image: url('/home/stage2_1.webp');"
		>
		</div>
		<Card.Root class="bg-linear-to-br from-white/8 to-white/2 bg-[#07091B]">
			<Card.Header>
				<Card.Title>Claim a part of a worldwide map with your Towers</Card.Title>
			</Card.Header>
			<Card.Content>
				<img src="/home/stage2_2.webp" alt="eSim" />
			</Card.Content>
		</Card.Root>
		<Card.Root class="bg-linear-to-br from-white/8 to-white/2 bg-[#07091B]">
			<Card.Header>
				<Card.Title>Earn even more $COMPANY_TOKEN tokens for claiming the area</Card.Title>
			</Card.Header>
			<Card.Content>
				<img src="/home/stage2_3.webp" alt="eSim" />
			</Card.Content>
		</Card.Root>
		<Card.Root class="bg-linear-to-br from-white/8 to-white/2 bg-[#07091B]">
			<Card.Header>
				<Card.Title>Use $COMPANY_TOKEN to buy an internet access or upgrade your towers to earn even more</Card.Title>
			</Card.Header>
			<Card.Content>
				<img src="/home/stage2_4.webp" alt="eSim" />
			</Card.Content>
		</Card.Root>
	</div>
</div>
