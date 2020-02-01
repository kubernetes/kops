# kops - Operaciones con Kubernetes

[![Build Status](https://travis-ci.org/kubernetes/kops.svg?branch=master)](https://travis-ci.org/kubernetes/kops) [![Go Report Card](https://goreportcard.com/badge/k8s.io/kops)](https://goreportcard.com/report/k8s.io/kops)  [![GoDoc Widget]][GoDoc]

[GoDoc]: https://pkg.go.dev/k8s.io/kops
[GoDoc Widget]: https://godoc.org/k8s.io/kops?status.svg


La forma más fácil de poner en marcha un cluster Kubernetes en producción.


## ¿Qué es kops?

Queremos pensar que es algo como `kubectl` para clusters.

`kops` ayuda a crear, destruir, mejorar y mantener un grado de producción, altamente
disponible, desde las líneas de comando de Kubernetes clusters. AWS (Amazon Web Services)
está oficialmente soportado actualmente, con GCE en soporte beta , y VMware vSphere
en alpha, y otras plataformas planeadas.


## ¿Puedo verlo en acción?

<p align="center">
  <a href="https://asciinema.org/a/97298">
  <img src="https://asciinema.org/a/97298.png" width="885"></image>
  </a>
</p>


## Lanzando un anfitrión de Kubernetes cluster en AWS o GCE

Para reproducir exactamente el demo anterior, visualizalo en el [tutorial](/docs/getting_started/aws.md) para
lanzar un anfitrión de Kubernetes cluster en AWS.

Para instalar un Kubernetes cluster en GCE por fabor siga esta [guide](/docs/getting_started/gce.md).


## Caracteristicas

* Automatiza el aprovisionamiento de Kubernetes clusters en [AWS](/docs/getting_started/aws.md) y [GCE](/docs/getting_started/gce.md)
* Un Despliegue Altamente Disponible (HA) Kubernetes Masters
* Construye en un modelo de estado sincronizado para **dry-runs** y **idempotency** automático
* Capacidad de generar [Terraform](/docs/terraform.md)
* Soporta un Kubernetes personalizado [add-ons](/docs/operations/addons.md)
* Línea de comando [autocompletion](/docs/cli/kops_completion.md)
* YAML Archivo de Manifiesto Basado en API [Configuration](/docs/manifests_and_customizing_via_api.md)
* [Templating](/docs/cluster_template.md) y ejecutar modos de simulacro para crear
 Manifiestos
* Escoge de ocho proveedores CNI diferentes [Networking](/docs/networking.md)
* Soporta Actualizarse desde [kube-up](/docs/upgrade_from_kubeup.md)
* Capacidad para añadir contenedores, como enganches, y archivos a nodos vía [cluster manifest](/docs/cluster_spec.md)


## Documentación

La documentación está en el directorio `/docs`, [and the index is here.](docs/README.md)


## Compatibilidad de Kubernetes con el Lanzamiento


### Soporte de la Versión Kubernetes

kops está destinado a ser compatible con versiones anteriores.  Siempre es recomendado utilizar la
última versión de kops con cualquier versión de Kubernetes que estés utilizando.  Siempre
utilize la última versión de kops.

Una excepción, en lo que respecta a la compatibilidad, kops soporta el equivalente a
un número de versión menor de Kubernetes.  Una versión menor es el segundo dígito en el
número de versión.  la versión de kops 1.8.0 tiene una versión menor de 8. La numeración
sigue la especificación de versión semántica, MAJOR.MINOR.PATCH.

Por ejemplo kops, 1.8.0 no soporta Kubernetes 1.9.2, pero kops 1.9.0
soporta Kubernetes 1.9.2 y versiones anteriores de Kubernetes. Sólo cuando coincide la versión
menor de kops, La versión menor de kubernetes hace que kops soporte oficialmente
el lanzamiento de kubernetes. kops no impide que un usuario instale versiones
no coincidentes de K8, pero las versiones de Kubernetes siempre requieren kops para instalar
versiones de componentes como docker, probado contra la versión
particular de Kubernetes.

#### Compatibilidad Matrix

| kops version | k8s 1.5.x | k8s 1.6.x | k8s 1.7.x | k8s 1.8.x | k8s 1.9.x |
|--------------|-----------|-----------|-----------|-----------|-----------|
| 1.9.x        | Y         | Y         | Y         | Y         | Y         |
| 1.8.x        | Y         | Y         | Y         | Y         | N         |
| 1.7.x        | Y         | Y         | Y         | N         | N         |
| 1.6.x        | Y         | Y         | N         | N         | N         |

Utilice la última versión de kops para todas las versiones de Kubernetes, con la advertencia de que las versiones más altas de Kubernetes no cuentan con el respaldo _oficial_ de kops.

### Cronograma de Lanzamiento de kops

Este proyecto no sigue el cronograma de lanzamiento de Kubernetes. `kops` tiene como objetivo
proporcionar una experiencia de instalación confiable para Kubernetes, y, por lo general, se lanza
aproximadamente un mes después de la publicación correspondiente de Kubernetes. Esta vez, permite que el proyecto Kubernetes resuelva los problemas que presenta la nueva versión y garantiza que podamos admitir las funciones más recientes. kops lanzará pre-lanzamientos alfa y beta para las personas que están ansiosas por probar la última versión de Kubernetes.
Utilice únicamente lanzamientos pre-GA kops en ambientes que puedan tolerar las peculiaridades de las nuevas versiones, e informe cualquier problema que surja.


## Instalación

### Requisito previo

`kubectl` es requerido, visualize [here](http://kubernetes.io/docs/user-guide/prereqs/).


### OSX desde Homebrew

```console
brew update && brew install kops
```

El binario `kops` también está disponible a través de nuestro [releases](https://github.com/kubernetes/kops/releases/latest).


### Linux

```console
curl -LO https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-linux-amd64
chmod +x kops-linux-amd64
sudo mv kops-linux-amd64 /usr/local/bin/kops
```


## Historial de Versiones

visualize el [releases](https://github.com/kubernetes/kops/releases) para más
información sobre cambios entre lanzamientos.


## Involucrarse y Contribuir

¿Estás interesado en contribuir con kops? Nosotros, los mantenedores y la comunidad,
nos encantaría sus sugerencias, contribuciones y ayuda.
Tenemos una guía de inicio rápido en [adding a feature](/docs/development/adding_a_feature.md). Además, se
puede contactar a los mantenedores en cualquier momento para obtener más información sobre
cómo involucrarse.
Con el interés de involucrar a más personas con kops, estamos comenzando a
etiquetar los problemas con `good-starter-issue`. Por lo general, se trata de problemas que tienen
un alcance menor, pero que son buenas maneras de familiarizarse con la base de código.

También alentamos a TODOS los participantes activos de la comunidad a actuar como si fueran
mantenedores, incluso si no tiene permisos de escritura "oficiales".Este es un
esfuerzo de la comunidad, estamos aquí para servir a la comunidad de Kubernetes.
Si tienes un interés activo y quieres involucrarte, ¡tienes verdadero poder!
No asuma que las únicas personas que pueden hacer cosas aquí son los "mantenedores".

También nos gustaría agregar más mantenedores "oficiales", así que
¡muéstranos lo que puedes hacer!


Lo que esto significa:

__Issues__
* Ayude a leer y clasifique los problemas, ayúdelo cuando sea posible.
* Señale los problemas que son duplicados, desactualizados, etc.
  - Incluso si no tiene permisos para etiquetar, tome nota y etiquete mantenedores (`/close`,`/dupe #127`).

__Pull Requests__
* Lee y revisa el código. Deja comentarios, preguntas y críticas (`/lgtm` ).
* Descargue, compile y ejecute el código y asegúrese de que las pruebas pasen (make test).
  - También verifique que la nueva característica parezca cuerda, siga los mejores patrones arquitectónicos e incluya pruebas.

Este repositorio usa los bots de Kubernetes.  Hay una lista completa de los comandos [aqui](https://go.k8s.io/bot-commands).


## Horas de Oficina

Los mantenedores de Kops reservaron una hora cada dos semanas para **horas de oficina** públicas. Los horarios de oficina se alojan en un [zoom video chat](https://zoom.us/my/k8ssigaws) los viernes en [5 pm UTC/12 noon ET/9 am US Pacific](http://www.worldtimebuddy.com/?pl=1&lid=100,5,8,12), en semanas impares numeradas. Nos esforzamos por conocer y ayudar a los programadores, ya sea trabajando en `kops` o interesados en conocer más sobre el proyecto.


### Temas Abiertos en Horas de Oficina

Incluye pero no limitado a:

- Ayuda y guía para aquellos que asisten, que están interesados en contribuir.
- Discuta el estado actual del proyecto kops, incluidas las versiones.
- Diseña estrategias para mover `kops` hacia adelante.
- Colabora sobre PRs abiertos y próximos.
- Presenta demos.

Esta vez se enfoca en los programadores, aunque nunca rechazaremos a un participante cortés. Pase por alto, incluso si nunca ha instalado Kops.

Le recomendamos que se comunique **de antemano** si planea asistir. Puedes unirte a cualquier sesión y no dudes en agregar un elemento a la [agenda](https://docs.google.com/document/d/12QkyL0FkNbWPcLFxxRGSPt_tNPBHbmni3YLY-lHny7E/edit) donde rastreamos notas en el horario de oficina.

Los horarios de oficina están alojados en una [Zoom](https://zoom.us/my/k8ssigaws) video conferencia, celebrada los viernes a las [5 pm UTC/12 noon ET/9 am US Pacific](http://www.worldtimebuddy.com/?pl=1&lid=100,5,8,12) cada otra semana impare numerada.

Puede verificar su número de semana utilizando:

```bash
date +%V
```

Los mantenedores y otros miembros de la comunidad están generalmente disponibles en [kubernetes slack](https://github.com/kubernetes/community/blob/master/communication.md#social-media) en [#kops](https://kubernetes.slack.com/messages/kops/), ¡así que ven y conversa con nosotros sobre cómo los kops pueden ser mejores para ti!


## GitHub Issues


### Errores

Si cree que ha encontrado un error, siga las instrucciones a continuación.

- Dedique una pequeña cantidad de tiempo a prestar la debida diligencia al rastreador de problemas. Tu problema puede ser un duplicado.
- Establezca la `-v 10` línea de comando y guarde la salida de los registros. Por favor pegue esto en su issue.
- Note the version of kops you are running (from `kops version`), and the command line options you are using.
- Abra un [new issue](https://github.com/kubernetes/kops/issues/new).
- Recuerde que los usuarios pueden estar buscando su issue en el futuro, por lo que debe darle un título significativo para ayudar a otros.
- No dude en comunicarse con la comunidad de kops en [kubernetes slack](https://github.com/kubernetes/community/blob/master/communication.md#social-media).


### Caracteristicas

También usamos el rastreador de problemas para rastrear características. Si tiene una idea para una función, o cree que puede ayudar a que los kops se vuelvan aún más impresionantes, siga los pasos a continuación.

- Abra un [new issue](https://github.com/kubernetes/kops/issues/new).
- Recuerde que los usuarios pueden estar buscando su issue en el futuro, por lo que debe darle un título significativo para ayudar a otros.
- Defina claramente el caso de uso, usando ejemplos concretos. P EJ: Escribo `esto` y kops hace `eso`.
- Algunas de nuestras características más grandes requerirán algún diseño. Si desea incluir un diseño técnico para su función, inclúyalo en el problema.
- Después de que la nueva característica sea bien comprendida, y el diseño acordado, podemos comenzar a codificar la característica. Nos encantaría que lo codificaras. Por lo tanto, abra una **WIP** *(trabajo en progreso)* solicitud de extracción, y que tenga una feliz codificación.
