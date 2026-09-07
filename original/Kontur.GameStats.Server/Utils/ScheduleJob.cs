using System;
using System.Threading;
using System.Threading.Tasks;

namespace Kontur.GameStats.Server.Utils {
    public class ScheduleJob {
        public static async Task ScheduleTask<T>(Action<T> action, TimeSpan interval, CancellationToken cancellationToken = default(CancellationToken)) where T : new() {
            await Task.Run(async () => {
                while (!cancellationToken.IsCancellationRequested) {
                    action(new T());
                    await Task.Delay(interval, cancellationToken);
                }
            }, cancellationToken);
        }

        public static async Task BackgroundTask<T>(Action<T> action, CancellationToken cancellationToken = default(CancellationToken)) where T : new() {
            await Task.Run(() => action(new T()), cancellationToken);
        }
    }
}
