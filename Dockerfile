FROM golang:1.16

ARG MODEL
# e.g. glove.42B.300d.zip

RUN echo $MODEL
RUN echo http://nlp.stanford.edu/data/glove.$MODEL.zip
