# run docker: docker run --rm -p 8787:8787 rocker/verse
# Go to: localhost:8787 (user/pass = rstudio/rstudio)
# in rstudio create new file "results.csv", copy&paste results
# execute:

library(ggplot2)
data = read.csv("results.csv", header=TRUE, sep=";")
df = data.frame(data)
ggplot(df, aes(x=vrp, y=delay, shape=policy, color=policy)) +
    geom_point() + scale_shape_manual(values=c(0, 2)) +
    xlab("Liczba pojazdów na minutę") +
    ylab("Średni czas przejazdu [s]") +
    ylim(0, NA) +
    xlim(0, NA)

# export result image
