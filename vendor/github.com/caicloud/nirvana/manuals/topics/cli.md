# CLI

CLI, or command-line interface is where you put the switches or entrances to your application's different functionalities. You can always adopt a micro-service approach and follow UNIX philosophy to build dedicated application for specific use-case, as you should always do. However you'll still need some good and consistent way to expose different options for people / system that run your application. Therefore it is very good practice to follow good convention and make your CLI unambiguous and first class API as well.

In Nirvana, we didn't build CLI framework from scratch and instead adopt [Cobra](https://github.com/spf13/cobra) as a default baseline, with some small addition just to boost your experience.

TODO: more details to follow
