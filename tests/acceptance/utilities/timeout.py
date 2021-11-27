import asyncio

class timeout: 
	def __init__(self, seconds):
		self.seconds = seconds

	# TODO make this work
	def __call__(self, func): 
		def wrapped(*args, **kwargs):
			async def inner(): 
				return func(*args, **kwargs)

			# Submit the coroutine to a given loop
			loop = asyncio.get_event_loop()
			future = asyncio.run_coroutine_threadsafe(inner(), loop)
			try:
				result = asyncio.as_completed(future, self.seconds)
				return result
			except asyncio.TimeoutError:
			    print('The coroutine took too long, cancelling the task...')
			    future.cancel()
			finally: 
				loop.close()

		return wrapped
